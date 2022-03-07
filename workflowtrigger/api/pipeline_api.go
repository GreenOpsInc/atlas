package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/cluster"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/event"
	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/kafkaclient"
	"github.com/greenopsinc/util/kubernetesclient"
	"github.com/greenopsinc/util/pipeline"
	"github.com/greenopsinc/util/team"
	"greenops.io/workflowtrigger/api/argo"
	"greenops.io/workflowtrigger/api/commanddelegator"
	"greenops.io/workflowtrigger/api/reposerver"
	"greenops.io/workflowtrigger/schemavalidation"
	"greenops.io/workflowtrigger/serializer"
)

const (
	orgNameField        string = "orgName"
	teamNameField       string = "teamName"
	pipelineNameField   string = "pipelineName"
	parentTeamNameField string = "parentTeamName"
	clusterNameField    string = "clusterName"
	namespaceField      string = "namespace"
	//Default val is ROOT_COMMIT
	revisionHashField string = "revisionHash"
	//Default val is LATEST_REVISION
	argoRevisionHashField string = "argoRevisionHash"
)

var dbOperator db.DbOperator
var kafkaClient kafkaclient.KafkaClient
var kubernetesClient kubernetesclient.KubernetesClient
var repoManagerApi reposerver.RepoManagerApi
var argoClusterApi argo.ArgoClusterApi
var argoRepoApi argo.ArgoRepoApi
var commandDelegatorApi commanddelegator.CommandDelegatorApi
var schemaValidator schemavalidation.RequestSchemaValidator

func createTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	parentTeamName := vars[parentTeamNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	key := db.MakeDbTeamKey(orgName, teamName)
	teamSchema := dbClient.FetchTeamSchema(key)
	if teamSchema.GetOrgName() == "" {
		newTeam := team.New(teamName, parentTeamName, orgName)
		dbClient.StoreValue(key, newTeam)
		addTeamToOrgList(newTeam.GetOrgName(), newTeam.GetTeamName(), dbClient)
		log.Printf("Created new team %s", newTeam.GetTeamName())
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "team already exists", http.StatusBadRequest)
}

func readTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	key := db.MakeDbTeamKey(orgName, teamName)
	teamSchema := dbClient.FetchTeamSchema(key)
	if teamSchema.GetOrgName() == "" {
		http.Error(w, "no team found", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teamSchema)
}

func listTeams(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	key := db.MakeDbListOfTeamsKey(orgName)
	listOfTeams := dbClient.FetchStringList(key)
	if listOfTeams == nil {
		listOfTeams = make([]string, 0)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(serializer.Serialize(listOfTeams)))
}

func deleteTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	key := db.MakeDbTeamKey(orgName, teamName)
	teamSchema := dbClient.FetchTeamSchema(key)
	for _, val := range teamSchema.GetPipelineSchemas() {
		if kubernetesClient.StoreGitCred(nil, db.MakeSecretName(orgName, teamName, val.GetPipelineName())) {
			http.Error(w, "kubernetes secret deletion did not work", http.StatusInternalServerError)
			return
		}
	}
	dbClient.StoreValue(key, nil)
	removeTeamFromOrgList(orgName, teamName, dbClient)
	w.WriteHeader(http.StatusOK)
}

func createPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	var gitRepo git.GitRepoSchema
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	gitRepo = git.UnmarshallGitRepoSchemaString(string(buf.Bytes()))
	key := db.MakeDbTeamKey(orgName, teamName)
	teamSchema := dbClient.FetchTeamSchema(key)
	if teamSchema.GetOrgName() == "" {
		http.Error(w, "no team schema found", http.StatusBadRequest)
		return
	}
	if teamSchema.GetPipelineSchema(pipelineName) != nil {
		http.Error(w, "pipeline already exists", http.StatusBadRequest)
		return
	}
	if !kubernetesClient.StoreGitCred(gitRepo.GetGitCred(), db.MakeSecretName(orgName, teamName, pipelineName)) {
		http.Error(w, "storing git credentials failed", http.StatusInternalServerError)
		return
	}
	if !repoManagerApi.CloneRepo(orgName, gitRepo) {
		http.Error(w, "cloning repository failed", http.StatusInternalServerError)
		return
	}

	err = argoRepoApi.CreateRepo(gitRepo.GetGitRepo(), gitRepo.GetGitCred())
	if err != nil {
		http.Error(w, fmt.Sprintf("cloning repository in argo failed: %s", err), http.StatusInternalServerError)
		return
	}

	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, git.GitRepoSchemaInfo{GitRepoUrl: gitRepo.GitRepo, PathToRoot: gitRepo.PathToRoot}, reposerver.RootCommit, string(argo.CreateAction), string(argo.ApplicationResource)) {
		repoManagerApi.DeleteRepo(gitRepo)
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	gitRepo.GetGitCred().Hide()

	teamSchema.AddPipeline(pipelineName, gitRepo)

	dbClient.StoreValue(key, teamSchema)
	w.WriteHeader(http.StatusOK)
}

func getPipelineEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	pipelineSchema := getPipeline(orgName, teamName, pipelineName, dbClient)
	if pipelineSchema == nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, git.GitRepoSchemaInfo{GitRepoUrl: pipelineSchema.GetGitRepoSchema().GitRepo, PathToRoot: pipelineSchema.GetGitRepoSchema().PathToRoot}, reposerver.RootCommit, string(argo.GetAction), string(argo.ApplicationResource)) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}
	bytesObj, _ := json.Marshal(pipelineSchema)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytesObj)
	w.WriteHeader(http.StatusOK)
}

func getPipeline(orgName string, teamName string, pipelineName string, dbClient db.DbClient) *pipeline.PipelineSchema {
	key := db.MakeDbTeamKey(orgName, teamName)
	teamSchema := dbClient.FetchTeamSchema(key)
	if teamSchema.GetOrgName() == "" {
		return nil
	}
	pipelineSchema := teamSchema.GetPipelineSchema(pipelineName)
	if pipelineSchema == nil {
		return nil
	}
	return pipelineSchema
}

func deletePipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	key := db.MakeDbTeamKey(orgName, teamName)
	teamSchema := dbClient.FetchTeamSchema(key)
	if teamSchema.GetOrgName() == "" {
		http.Error(w, "team does not exist", http.StatusBadRequest)
		return
	}
	if teamSchema.GetPipelineSchema(pipelineName) == nil {
		http.Error(w, "pipeline does not exist", http.StatusBadRequest)
		return
	}
	gitRepo := teamSchema.GetPipelineSchema(pipelineName).GetGitRepoSchema()
	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, git.GitRepoSchemaInfo{GitRepoUrl: gitRepo.GitRepo, PathToRoot: gitRepo.PathToRoot}, reposerver.RootCommit, string(argo.DeleteAction), string(argo.ApplicationResource)) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}
	kubernetesClient.StoreGitCred(nil, db.MakeSecretName(orgName, teamName, pipelineName))

	teamSchema.RemovePipeline(pipelineName)

	if !repoManagerApi.DeleteRepo(gitRepo) {
		http.Error(w, "repository could not be deleted", http.StatusInternalServerError)
		return
	}
	clearPipelineData(orgName, teamName, pipelineName)

	dbClient.StoreValue(key, teamSchema)
	w.WriteHeader(http.StatusOK)
	return
}

func clearPipelineData(orgName string, teamName string, pipelineName string) {
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	prefix := fmt.Sprintf("%s-%s-%s", orgName, teamName, pipelineName)
	dbClient.DeleteByPrefix(prefix)
}

func updatePipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	key := db.MakeDbTeamKey(orgName, teamName)

	var gitRepoUpd git.GitRepoSchema
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	gitRepoUpd = git.UnmarshallGitRepoSchemaString(string(buf.Bytes()))

	teamSchema := dbClient.FetchTeamSchema(key)
	if teamSchema.GetOrgName() == "" {
		http.Error(w, "team does not exist", http.StatusBadRequest)
		return
	}
	if teamSchema.GetPipelineSchema(pipelineName) == nil {
		http.Error(w, "pipeline does not exist", http.StatusBadRequest)
		return
	}

	gitRepo := teamSchema.GetPipelineSchema(pipelineName).GetGitRepoSchema()
	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, git.GitRepoSchemaInfo{GitRepoUrl: gitRepo.GitRepo, PathToRoot: gitRepo.PathToRoot}, reposerver.RootCommit, string(argo.UpdateAction), string(argo.ApplicationResource)) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
	}
	kubernetesClient.StoreGitCred(gitRepoUpd.GetGitCred(), db.MakeSecretName(orgName, teamName, pipelineName))

	teamSchema.UpdatePipeline(pipelineName, gitRepoUpd)

	if !repoManagerApi.UpdateRepo(orgName, gitRepo, gitRepoUpd) {
		http.Error(w, "updating the repository failed", http.StatusInternalServerError)
		return
	}
	dbClient.StoreValue(key, teamSchema)
	w.WriteHeader(http.StatusOK)
	return
}

func syncPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	revisionHash := vars[revisionHashField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	var gitRepo git.GitRepoSchema
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	gitRepo = git.UnmarshallGitRepoSchemaString(string(buf.Bytes()))
	currentRevisionHash := repoManagerApi.SyncRepo(gitRepo)
	if currentRevisionHash == "" {
		http.Error(w, "syncing repository failed", http.StatusBadRequest)
		return
	}

	if revisionHash == reposerver.RootCommit {
		revisionHash = currentRevisionHash
	}

	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, git.GitRepoSchemaInfo{GitRepoUrl: gitRepo.GitRepo, PathToRoot: gitRepo.PathToRoot}, revisionHash,
		string(argo.SyncAction), string(argo.ApplicationResource)) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	triggerEvent := event.NewPipelineTriggerEvent(orgName, teamName, pipelineName)
	triggerEvent.(*event.PipelineTriggerEvent).RevisionHash = revisionHash
	payload := serializer.Serialize(triggerEvent)
	generateEvent(payload)
	w.WriteHeader(http.StatusOK)
}

func runSubPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	stepName := vars[stepNameField]
	revisionHash := vars[revisionHashField]
	var gitRepo git.GitRepoSchema
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	gitRepo = git.UnmarshallGitRepoSchemaString(string(buf.Bytes()))
	currentRevisionHash := repoManagerApi.SyncRepo(gitRepo)
	if currentRevisionHash == "" {
		http.Error(w, "syncing repository failed", http.StatusBadRequest)
		return
	}

	if revisionHash == reposerver.RootCommit {
		revisionHash = currentRevisionHash
	}

	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, git.GitRepoSchemaInfo{GitRepoUrl: gitRepo.GitRepo, PathToRoot: gitRepo.PathToRoot}, revisionHash,
		string(argo.OverrideAction), string(argo.ApplicationResource),
		string(argo.SyncAction), string(argo.ApplicationResource)) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	triggerEvent := event.NewPipelineTriggerEvent(orgName, teamName, pipelineName)
	triggerEvent.(*event.PipelineTriggerEvent).StepName = stepName
	triggerEvent.(*event.PipelineTriggerEvent).RevisionHash = revisionHash
	payload := serializer.Serialize(triggerEvent)
	generateEvent(payload)
	w.WriteHeader(http.StatusOK)
}

func forceDeploy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	stepName := vars[stepNameField]
	pipelineRevisionHash := vars[revisionHashField]
	argoRevisionHash := vars[argoRevisionHashField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	var gitRepo git.GitRepoSchema
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	gitRepo = git.UnmarshallGitRepoSchemaString(string(buf.Bytes()))
	currentRevisionHash := repoManagerApi.SyncRepo(gitRepo)
	if currentRevisionHash == "" {
		http.Error(w, "syncing repository failed", http.StatusBadRequest)
		return
	}

	if pipelineRevisionHash == reposerver.RootCommit {
		pipelineRevisionHash = currentRevisionHash
	}

	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, git.GitRepoSchemaInfo{GitRepoUrl: gitRepo.GitRepo, PathToRoot: gitRepo.PathToRoot}, pipelineRevisionHash,
		string(argo.OverrideAction), string(argo.ApplicationResource),
		string(argo.SyncAction), string(argo.ApplicationResource)) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	applicationPayload, clusterName := schemaValidator.GetStepApplicationPayloadWithoutPipelineData(orgName, teamName, git.GitRepoSchemaInfo{GitRepoUrl: gitRepo.GitRepo, PathToRoot: gitRepo.PathToRoot}, pipelineRevisionHash, stepName)

	clusterSchema := dbClient.FetchClusterSchema(db.MakeDbClusterKey(orgName, clusterName))
	emptyStruct := cluster.ClusterSchema{}
	if clusterSchema == emptyStruct {
		http.Error(w, "Cluster does not exist", http.StatusBadRequest)
	}
	if clusterSchema.NoDeploy != nil {
		http.Error(w, "No deploy is enabled for this cluster, the request will be blocked", http.StatusBadRequest)
	}

	deployRequest := &clientrequest.ClientDeployRequest{
		ClientRequestEventMetadata: clientrequest.ClientRequestEventMetadata{
			OrgName:      orgName,
			TeamName:     teamName,
			PipelineName: pipelineName,
			PipelineUvn:  uuid.New().String(),
			StepName:     stepName,
		},
		ResponseEventType: "",
		DeployType:        "DeployArgoRequest",
		RevisionHash:      argoRevisionHash,
		Payload:           applicationPayload,
	}

	payload := serializer.Serialize(clientrequest.ClientRequestPacket{
		RetryCount:    0,
		Namespace:     schemaValidator.GetArgoApplicationNamespace(applicationPayload),
		ClientRequest: deployRequest,
	})
	dbClient.InsertValueInTransactionlessList(db.MakeClientRequestQueueKey(orgName, clusterName), payload)
	w.WriteHeader(http.StatusOK)
}

func aggregatePipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]
	namespace := vars[namespaceField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	if clusterName == "" || namespace == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	requestId := commandDelegatorApi.SendNotification(orgName, clusterName, &clientrequest.ClientAggregateRequest{
		ClusterName: clusterName,
		Namespace:   namespace,
	})

	notification := getNotification(requestId, dbClient)
	if !notification.Successful {
		http.Error(w, notification.Body, http.StatusInternalServerError)
		return
	}
	//bytesObj, _ := json.Marshal(notification.Body)
	w.Write([]byte(notification.Body))
	w.Header().Set("Content-Type", "application/json")
	//w.WriteHeader(http.StatusOK)
}

func labelPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	var resourceList clientrequest.GvkGroupRequest
	var labelRequest clientrequest.ClientLabelRequest

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	err = json.Unmarshal(buf.Bytes(), &resourceList)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	labelRequest.OrgName = orgName
	labelRequest.TeamName = teamName
	labelRequest.PipelineName = pipelineName
	labelRequest.GvkResourceList = resourceList

	requestId := commandDelegatorApi.SendNotification(orgName, clusterName, &labelRequest)

	notification := getNotification(requestId, dbClient)
	if !notification.Successful {
		http.Error(w, notification.Body, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteByLabel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	namespace := vars[namespaceField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	pipelineSchema := getPipeline(orgName, teamName, pipelineName, dbClient)
	var repoInfo git.GitRepoSchemaInfo
	repoInfo.GitRepoUrl = pipelineSchema.GetGitRepoSchema().GitRepo
	repoInfo.PathToRoot = pipelineSchema.GetGitRepoSchema().PathToRoot

	if pipelineSchema == nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, repoInfo, reposerver.RootCommit, string(argo.DeleteAction), string(argo.ApplicationResource)) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	var request clientrequest.ClientDeleteByLabelRequest
	request.TeamName = teamName
	request.PipelineName = pipelineName
	request.Namespace = namespace

	requestId := commandDelegatorApi.SendNotification(orgName, clusterName, &request)
	notification := getNotification(requestId, dbClient)
	if !notification.Successful {
		http.Error(w, notification.Body, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getPipelineClusterNamespaceCombinations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	pipelineSchema := getPipeline(orgName, teamName, pipelineName, dbClient)
	var repoInfo git.GitRepoSchemaInfo
	repoInfo.GitRepoUrl = pipelineSchema.GetGitRepoSchema().GitRepo
	repoInfo.PathToRoot = pipelineSchema.GetGitRepoSchema().PathToRoot

	if pipelineSchema == nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	groups := schemaValidator.GetClusterNamespaceCombinations(orgName, teamName, repoInfo, reposerver.RootCommit)
	bytesObj, _ := json.Marshal(groups)
	w.Write(bytesObj)
	w.Header().Set("Content-Type", "application/json")
}

func generateEventEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	//orgName := vars[orgNameField]
	clusterName := vars[clusterNameField]
	if !schemaValidator.VerifyRbac(argo.UpdateAction, argo.ClusterResource, clusterName) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	generateEvent(buf.String())
	w.WriteHeader(http.StatusOK)
	return
}

func generateEvent(event string) {
	err := kafkaClient.SendMessage(event)
	if err != nil {
		panic(err)
	}
}

func generateNotification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestId := vars["requestId"]
	dbClient := dbOperator.GetClient()
	defer dbClient.Close()

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	key := db.MakeDbNotificationKey(requestId)
	dbClient.StoreValue(key, buf.String())
	w.WriteHeader(http.StatusOK)
	return
}

func removeTeamFromOrgList(orgName string, teamName string, dbClient db.DbClient) {
	key := db.MakeDbListOfTeamsKey(orgName)
	listOfTeams := dbClient.FetchStringList(key)
	if listOfTeams == nil {
		listOfTeams = make([]string, 0)
	}
	for idx, val := range listOfTeams {
		if val == teamName {
			listOfTeams = append(listOfTeams[:idx], listOfTeams[idx+1:]...)
		}
	}
	dbClient.StoreValue(key, listOfTeams)
}

func addTeamToOrgList(orgName string, teamName string, dbClient db.DbClient) {
	key := db.MakeDbListOfTeamsKey(orgName)
	listOfTeams := dbClient.FetchStringList(key)
	if listOfTeams == nil {
		listOfTeams = make([]string, 0)
	}
	for _, val := range listOfTeams {
		if val == teamName {
			return
		}
	}
	listOfTeams = append(listOfTeams, teamName)
	dbClient.StoreValue(key, listOfTeams)
}

func getNotification(requestId string, dbClient db.DbClient) clientrequest.Notification {
	key := db.MakeDbNotificationKey(requestId)
	time.Sleep(5 * time.Second)
	var notification clientrequest.Notification
	emptyStruct := clientrequest.Notification{}
	for i := 0; i < 60; i++ {
		notification = dbClient.FetchNotification(key)
		if notification != emptyStruct {
			dbClient.StoreValue(key, nil)
			return notification
		}
		time.Sleep(1 * time.Second)
	}
	return clientrequest.Notification{
		Successful: false,
		Body:       "error: The request response could not be found",
	}
}

func InitPipelineTeamEndpoints(r *mux.Router) {
	r.HandleFunc("/team/{orgName}", listTeams).Methods("GET")
	r.HandleFunc("/team/{orgName}/{parentTeamName}/{teamName}", createTeam).Methods("POST")
	r.HandleFunc("/team/{orgName}/{teamName}", readTeam).Methods("GET")
	r.HandleFunc("/team/{orgName}/{teamName}", deleteTeam).Methods("DELETE")
	r.HandleFunc("/pipeline/{orgName}/{teamName}/{pipelineName}", createPipeline).Methods("POST")
	r.HandleFunc("/pipeline/{orgName}/{teamName}/{pipelineName}", updatePipeline).Methods("PUT")
	r.HandleFunc("/pipeline/{orgName}/{teamName}/{pipelineName}", getPipelineEndpoint).Methods("GET")
	r.HandleFunc("/pipeline/{orgName}/{teamName}/{pipelineName}", deletePipeline).Methods("DELETE")
	r.HandleFunc("/sync/{orgName}/{teamName}/{pipelineName}/{revisionHash}", syncPipeline).Methods("POST")
	r.HandleFunc("/sync/{orgName}/{teamName}/{pipelineName}/{revisionHash}/{stepName}", runSubPipeline).Methods("POST")
	r.HandleFunc("/force/{orgName}/{teamName}/{pipelineName}/{revisionHash}/{stepName}/{argoRevisionHash}", forceDeploy).Methods("POST")
	r.HandleFunc("/client/generateNotification/{requestId}", generateNotification).Methods("POST")
	r.HandleFunc("/client/{orgName}/{clusterName}/generateEvent", generateEventEndpoint).Methods("POST")
	r.HandleFunc("/combinations/{orgName}/{teamName}/{pipelineName}", getPipelineClusterNamespaceCombinations).Methods("GET")
	r.HandleFunc("/aggregate/{orgName}/{clusterName}/{namespace}", aggregatePipeline).Methods("GET")
	r.HandleFunc("/label/{orgName}/{clusterName}/{teamName}/{pipelineName}", labelPipeline).Methods("POST")
	r.HandleFunc("/clean/{orgName}/{clusterName}/{teamName}/{pipelineName}/{namespace}", deleteByLabel).Methods("POST")
}

func InitClients(dbOperatorCopy db.DbOperator, kafkaClientCopy kafkaclient.KafkaClient, kubernetesClientCopy kubernetesclient.KubernetesClient, repoManagerApiCopy reposerver.RepoManagerApi, argoClusterApiCopy argo.ArgoClusterApi, argoRepoApiCopy argo.ArgoRepoApi, commandDelegatorApiCopy commanddelegator.CommandDelegatorApi, schemaValidatorCopy schemavalidation.RequestSchemaValidator) {
	dbOperator = dbOperatorCopy
	kafkaClient = kafkaClientCopy
	kubernetesClient = kubernetesClientCopy
	repoManagerApi = repoManagerApiCopy
	argoClusterApi = argoClusterApiCopy
	argoRepoApi = argoRepoApiCopy
	commandDelegatorApi = commandDelegatorApiCopy
	schemaValidator = schemaValidatorCopy
}
