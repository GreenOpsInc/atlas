package api

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"greenops.io/workflowtrigger/api/argoauthenticator"
	"greenops.io/workflowtrigger/api/commanddelegator"
	"greenops.io/workflowtrigger/api/reposerver"
	"greenops.io/workflowtrigger/db"
	"greenops.io/workflowtrigger/kafka"
	"greenops.io/workflowtrigger/kubernetesclient"
	"greenops.io/workflowtrigger/schemavalidation"
	"greenops.io/workflowtrigger/util/clientrequest"
	"greenops.io/workflowtrigger/util/cluster"
	"greenops.io/workflowtrigger/util/event"
	"greenops.io/workflowtrigger/util/git"
	"greenops.io/workflowtrigger/util/pipeline"
	"greenops.io/workflowtrigger/util/serializer"
	"greenops.io/workflowtrigger/util/team"
	"log"
	"net/http"
	"time"
)

const (
	orgNameField        string = "orgName"
	teamNameField       string = "teamName"
	pipelineNameField   string = "pipelineName"
	parentTeamNameField string = "parentTeamName"
	clusterNameField    string = "clusterName"
	//Default val is ROOT_COMMIT
	revisionHashField string = "revisionHash"
	//Default val is LATEST_REVISION
	argoRevisionHashField string = "argoRevisionHash"
)

var dbClient db.DbClient
var kafkaClient kafka.KafkaClient
var kubernetesClient kubernetesclient.KubernetesClient
var repoManagerApi reposerver.RepoManagerApi
var commandDelegatorApi commanddelegator.CommandDelegatorApi
var schemaValidator schemavalidation.RequestSchemaValidator

func createTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	parentTeamName := vars[parentTeamNameField]
	key := db.MakeDbTeamKey(orgName, teamName)
	teamSchema := dbClient.FetchTeamSchema(key)
	if teamSchema.GetOrgName() == "" {
		newTeam := team.New(teamName, parentTeamName, orgName)
		dbClient.StoreValue(key, newTeam)
		addTeamToOrgList(newTeam.GetOrgName(), newTeam.GetTeamName())
		log.Printf("Created new team %s", newTeam.GetTeamName())
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func readTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	key := db.MakeDbTeamKey(orgName, teamName)
	teamSchema := dbClient.FetchTeamSchema(key)
	if teamSchema.GetOrgName() == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teamSchema)
}

func listTeams(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
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
	key := db.MakeDbTeamKey(orgName, teamName)
	teamSchema := dbClient.FetchTeamSchema(key)
	for _, val := range teamSchema.GetPipelineSchemas() {
		if kubernetesClient.StoreGitCred(nil, db.MakeSecretName(orgName, teamName, val.GetPipelineName())) {
			http.Error(w, "Kubernetes secret deletion did not work", http.StatusInternalServerError)
			return
		}
	}
	dbClient.StoreValue(key, nil)
	removeTeamFromOrgList(orgName, teamName)
	w.WriteHeader(http.StatusOK)
}

func createPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	var gitRepo git.GitRepoSchema
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	gitRepo = git.UnmarshallGitRepoSchemaString(string(buf.Bytes()))
	key := db.MakeDbTeamKey(orgName, teamName)
	teamSchema := dbClient.FetchTeamSchema(key)
	if teamSchema.GetOrgName() == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if teamSchema.GetPipelineSchema(pipelineName) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !kubernetesClient.StoreGitCred(gitRepo.GetGitCred(), db.MakeSecretName(orgName, teamName, pipelineName)) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !repoManagerApi.CloneRepo(orgName, gitRepo) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, gitRepo.GitRepo, reposerver.RootCommit, string(argoauthenticator.CreateAction), string(argoauthenticator.ApplicationResource)) {
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
	pipelineSchema := getPipeline(orgName, teamName, pipelineName)
	if pipelineSchema == nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, pipelineSchema.GetGitRepoSchema().GitRepo, reposerver.RootCommit, string(argoauthenticator.GetAction), string(argoauthenticator.ApplicationResource)) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}
	bytesObj, _ := json.Marshal(pipelineSchema)
	w.Write(bytesObj)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func getPipeline(orgName string, teamName string, pipelineName string) *pipeline.PipelineSchema {
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
	key := db.MakeDbTeamKey(orgName, teamName)
	teamSchema := dbClient.FetchTeamSchema(key)
	if teamSchema.GetOrgName() == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if teamSchema.GetPipelineSchema(pipelineName) == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, teamSchema.GetPipelineSchema(pipelineName).GetGitRepoSchema().GitRepo, reposerver.RootCommit, string(argoauthenticator.DeleteAction), string(argoauthenticator.ApplicationResource)) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}
	kubernetesClient.StoreGitCred(nil, db.MakeSecretName(orgName, teamName, pipelineName))
	gitRepo := teamSchema.GetPipelineSchema(pipelineName).GetGitRepoSchema()

	teamSchema.RemovePipeline(pipelineName)

	if !repoManagerApi.DeleteRepo(gitRepo) {
		w.WriteHeader(http.StatusInternalServerError)
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
	var gitRepo git.GitRepoSchema
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	gitRepo = git.UnmarshallGitRepoSchemaString(string(buf.Bytes()))
	if !repoManagerApi.SyncRepo(gitRepo) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, gitRepo.GitRepo, revisionHash,
		string(argoauthenticator.SyncAction), string(argoauthenticator.ApplicationResource),
		string(argoauthenticator.SyncAction), string(argoauthenticator.ClusterResource)) {
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	gitRepo = git.UnmarshallGitRepoSchemaString(string(buf.Bytes()))
	if !repoManagerApi.SyncRepo(gitRepo) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, gitRepo.GitRepo, revisionHash,
		string(argoauthenticator.OverrideAction), string(argoauthenticator.ApplicationResource),
		string(argoauthenticator.SyncAction), string(argoauthenticator.ApplicationResource),
		string(argoauthenticator.SyncAction), string(argoauthenticator.ClusterResource)) {
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
	var gitRepo git.GitRepoSchema
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	gitRepo = git.UnmarshallGitRepoSchemaString(string(buf.Bytes()))
	if !repoManagerApi.SyncRepo(gitRepo) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, gitRepo.GitRepo, pipelineRevisionHash,
		string(argoauthenticator.OverrideAction), string(argoauthenticator.ApplicationResource),
		string(argoauthenticator.SyncAction), string(argoauthenticator.ApplicationResource),
		string(argoauthenticator.SyncAction), string(argoauthenticator.ClusterResource)) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	applicationPayload, clusterName := schemaValidator.GetStepApplicationPayload(orgName, teamName, gitRepo.GitRepo, pipelineRevisionHash, stepName)

	clusterSchema := dbClient.FetchClusterSchema(db.MakeDbClusterKey(orgName, clusterName))
	emptyStruct := cluster.ClusterSchema{}
	if clusterSchema == emptyStruct {
		http.Error(w, "Cluster does not exist", http.StatusBadRequest)
	}
	if clusterSchema.NoDeploy != nil {
		http.Error(w, "No deploy is enabled for this cluster, the request will be blocked", http.StatusBadRequest)
	}

	deployRequest := clientrequest.ClientDeployRequest{
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

func generateEventEndpoint(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	//orgName := vars[orgNameField]
	//clusterName := vars[clusterNameField]
	//if !schemaValidator.VerifyRbac(argoauthenticator.UpdateAction, argoauthenticator.ClusterResource, clusterName) {
	//	http.Error(w, "Not enough permissions", http.StatusForbidden)
	//	return
	//}
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key := db.MakeDbNotificationKey(requestId)
	dbClient.StoreValue(key, buf.String())
	w.WriteHeader(http.StatusOK)
	return
}

func removeTeamFromOrgList(orgName string, teamName string) {
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

func addTeamToOrgList(orgName string, teamName string) {
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

func getNotification(requestId string) clientrequest.Notification {
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
	r.HandleFunc("/pipeline/{orgName}/{teamName}/{pipelineName}", getPipelineEndpoint).Methods("GET")
	r.HandleFunc("/pipeline/{orgName}/{teamName}/{pipelineName}", deletePipeline).Methods("DELETE")
	r.HandleFunc("/sync/{orgName}/{teamName}/{pipelineName}/{revisionHash}", syncPipeline).Methods("POST")
	r.HandleFunc("/sync/{orgName}/{teamName}/{pipelineName}/{revisionHash}/{stepName}", runSubPipeline).Methods("POST")
	r.HandleFunc("/force/{orgName}/{teamName}/{pipelineName}/{revisionHash}/{stepName}/{argoRevisionHash}", forceDeploy).Methods("POST")
	r.HandleFunc("/client/generateNotification/{requestId}", generateNotification).Methods("POST")
	r.HandleFunc("/client/{orgName}/{clusterName}/generateEvent", generateEventEndpoint).Methods("POST")
}

func InitClients(dbClientCopy db.DbClient, kafkaClientCopy kafka.KafkaClient, kubernetesClientCopy kubernetesclient.KubernetesClient, repoManagerApiCopy reposerver.RepoManagerApi, commandDelegatorApiCopy commanddelegator.CommandDelegatorApi, schemaValidatorCopy schemavalidation.RequestSchemaValidator) {
	dbClient = dbClientCopy
	kafkaClient = kafkaClientCopy
	kubernetesClient = kubernetesClientCopy
	repoManagerApi = repoManagerApiCopy
	commandDelegatorApi = commandDelegatorApiCopy
	schemaValidator = schemaValidatorCopy
}
