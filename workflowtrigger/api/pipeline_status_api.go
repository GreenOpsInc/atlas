package api

import (
	"fmt"
	"github.com/gorilla/mux"
	"greenops.io/workflowtrigger/api/argoauthenticator"
	"greenops.io/workflowtrigger/api/reposerver"
	"greenops.io/workflowtrigger/db"
	"greenops.io/workflowtrigger/pipelinestatus"
	"greenops.io/workflowtrigger/util/auditlog"
	"greenops.io/workflowtrigger/util/serializer"
	"math"
	"net/http"
	"strconv"
)

const (
	stepNameField    string = "stepName"
	pipelineUvnField string = "pipelineUvn"
	countField       string = "count"
)

func getStepLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	stepName := vars[stepNameField]

	pipelineSchema := getPipeline(orgName, teamName, pipelineName)
	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, pipelineSchema.GetGitRepoSchema().GitRepo, reposerver.RootCommit,
		string(argoauthenticator.GetAction), string(argoauthenticator.ApplicationResource)) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	count, err := strconv.Atoi(vars[countField])
	if err != nil {
		http.Error(w, "Count variable could not be deserialized", http.StatusBadRequest)
		return
	}
	key := db.MakeDbStepKey(orgName, teamName, pipelineName, stepName)
	increments := int(math.Ceil(float64(db.LogIncrement / count)))
	var logList []auditlog.Log
	var fetchedLogList []auditlog.Log
	for idx := 0; idx < increments; idx++ {
		fetchedLogList = dbClient.FetchLogList(key, idx)
		if idx == increments-1 {
			difference := count - ((increments - 1) * db.LogIncrement)
			logList = append(logList, fetchedLogList[0:int(math.Min(float64(difference), float64(len(fetchedLogList))))]...)
		} else {
			logList = append(logList, fetchedLogList...)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(serializer.Serialize(logList)))
}

func getPipelineUvns(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]

	pipelineSchema := getPipeline(orgName, teamName, pipelineName)
	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, pipelineSchema.GetGitRepoSchema().GitRepo, reposerver.RootCommit,
		string(argoauthenticator.GetAction), string(argoauthenticator.ApplicationResource)) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	count, err := strconv.Atoi(vars[countField])
	if err != nil {
		http.Error(w, "Count variable could not be deserialized", http.StatusBadRequest)
		return
	}
	key := db.MakeDbPipelineInfoKey(orgName, teamName, pipelineName)
	increments := int(math.Ceil(float64(db.LogIncrement / count)))
	var pipelineUvnList []string
	var fetchedPipelineUvnList []string
	for idx := 0; idx < increments; idx++ {
		for _, pipelineInfo := range dbClient.FetchPipelineInfoList(key, idx) {
			fetchedPipelineUvnList = append(fetchedPipelineUvnList, pipelineInfo.PipelineUvn)
		}
		if idx == increments-1 {
			difference := count - ((increments - 1) * db.LogIncrement)
			pipelineUvnList = append(pipelineUvnList, fetchedPipelineUvnList[0:int(math.Min(float64(difference), float64(len(fetchedPipelineUvnList))))]...)
		} else {
			pipelineUvnList = append(pipelineUvnList, fetchedPipelineUvnList...)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(serializer.Serialize(pipelineUvnList)))
}

func getPipelineStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	pipelineUvn := vars[pipelineUvnField]

	pipelineSchema := getPipeline(orgName, teamName, pipelineName)
	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, pipelineSchema.GetGitRepoSchema().GitRepo, reposerver.RootCommit,
		string(argoauthenticator.GetAction), string(argoauthenticator.ApplicationResource)) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	status := pipelinestatus.New()
	//Get pipeline UVN if not specified
	pipelineInfoKey := db.MakeDbPipelineInfoKey(orgName, teamName, pipelineName)
	var pipelineInfo auditlog.PipelineInfo
	if pipelineUvn == "LATEST" {
		pipelineInfo = dbClient.FetchLatestPipelineInfo(pipelineInfoKey)
		pipelineUvn = pipelineInfo.PipelineUvn
	} else {
		logIncrement := 0
		pipelineInfoList := dbClient.FetchPipelineInfoList(pipelineInfoKey, logIncrement)
		idx := 0
		for idx < len(pipelineInfoList) {
			if pipelineInfoList[idx].PipelineUvn == pipelineUvn {
				pipelineInfo = pipelineInfoList[idx]
				break
			}
			idx++
			if idx == len(pipelineInfoList) {
				logIncrement++
				pipelineInfoList = dbClient.FetchPipelineInfoList(pipelineInfoKey, logIncrement)
				idx = 0
			}
		}
	}
	if pipelineInfo.PipelineUvn == "" && len(pipelineInfo.Errors) == 0 {
		http.Error(w, "No pipeline runs exist with the requested UVN", http.StatusBadRequest)
		return
	}
	steps := pipelineInfo.StepList
	for _, step := range steps {
		//Get pipeline UVN if not specified
		logKey := db.MakeDbStepKey(orgName, teamName, pipelineName, step)
		var log auditlog.Log

		//TODO: This iteration is in enough places where it should be extracted as a dbClient method
		//Get most recent log (deployment or remediation) with desired pipeline UVN
		logIncrement := 0
		logList := dbClient.FetchLogList(logKey, logIncrement)
		idx := 0
		for idx < len(logList) {
			if logList[idx].GetPipelineUniqueVersionNumber() == pipelineUvn {
				log = logList[idx]
				break
			}
			idx++
			if idx == len(logList) {
				logIncrement++
				logList = dbClient.FetchLogList(logKey, logIncrement)
				idx = 0
			}
		}

		if log == nil {
			status.MarkIncomplete()
			continue
		}
		if tempDeploymentLog, ok := log.(*auditlog.DeploymentLog); ok {
			if tempDeploymentLog.GetStatus() == auditlog.Progressing {
				status.AddProgressingStep(step)
				continue
			}
		}
		if tempDeploymentLog, ok := log.(*auditlog.DeploymentLog); ok {
			if tempDeploymentLog.GetStatus() == auditlog.Cancelled {
				status.MarkCancelled()
				if tempDeploymentLog.BrokenTest != "" {
					status.AddFailedDeploymentLog(*tempDeploymentLog, step)
				}
				continue
			}
		}
		//Determines if the step is stable
		status.AddLatestLog(log)
		//Determines if the step is complete
		if tempDeploymentLog, ok := log.(*auditlog.DeploymentLog); ok {
			status.AddLatestDeploymentLog(*tempDeploymentLog, step)
		} else {
			log = nil
			for idx < len(logList) {
				if logList[idx].GetPipelineUniqueVersionNumber() == pipelineUvn {
					if _, ok = logList[idx].(*auditlog.DeploymentLog); ok {
						log = logList[idx]
						status.AddLatestDeploymentLog(*(log.(*auditlog.DeploymentLog)), step)
						break
					}
				}
				idx++
				if idx == len(logList) {
					logIncrement++
					logList = dbClient.FetchLogList(logKey, logIncrement)
					idx = 0
				}
			}
			if log == nil {
				status.MarkIncomplete()
				continue
			}
		}

		// was rollback deployment, need to get failure logs of initial deployment
		if log.GetUniqueVersionInstance() != 0 {
			logIncrement = 0
			logList = dbClient.FetchLogList(logKey, logIncrement)
			idx = 0
			for idx < len(logList) {
				if logList[idx].GetUniqueVersionInstance() == 0 {
					if tempDeploymentLog, ok := logList[idx].(*auditlog.DeploymentLog); ok {
						status.AddFailedDeploymentLog(*tempDeploymentLog, step)
						break
					}
				}
				idx++
				if idx == len(logList) {
					logIncrement++
					logList = dbClient.FetchLogList(logKey, logIncrement)
					idx = 0
				}
			}
		}
	}
	//Get additional "floating" errors. Largely related to processing
	if !(pipelineInfo.PipelineUvn == "" && len(pipelineInfo.Errors) == 0) {
		for idx, err := range pipelineInfo.Errors {
			status.AddFailedDeploymentLog(auditlog.DeploymentLog{
				DeploymentComplete: false,
				BrokenTest:         fmt.Sprintf("Processing error %d", idx),
				BrokenTestLog:      err,
			}, "Unknown")
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(serializer.Serialize(status)))
}

func cancelLatestPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]

	pipelineSchema := getPipeline(orgName, teamName, pipelineName)
	if !schemaValidator.ValidateSchemaAccess(orgName, teamName, pipelineSchema.GetGitRepoSchema().GitRepo, reposerver.RootCommit,
		string(argoauthenticator.SyncAction), string(argoauthenticator.ApplicationResource),
		string(argoauthenticator.SyncAction), string(argoauthenticator.ClusterResource)) {
		http.Error(w, "Not enough permissions", http.StatusForbidden)
		return
	}

	latestUvn := ""
	steps := dbClient.FetchLatestPipelineInfo(db.MakeDbPipelineInfoKey(orgName, teamName, pipelineName)).StepList
	for _, stepName := range steps {
		key := db.MakeDbStepKey(orgName, teamName, pipelineName, stepName)
		latestLog := dbClient.FetchLatestLog(key)
		if len(latestUvn) == 0 {
			//Step list is ordered, so if the very first log is nonexistant, it can't have deployed anywhere else
			if latestLog == nil {
				w.WriteHeader(http.StatusOK)
				return
			}
			latestUvn = latestLog.GetPipelineUniqueVersionNumber()
		}
		if latestLog == nil || latestLog.GetPipelineUniqueVersionNumber() != latestUvn {
			newCancelledLog := auditlog.InitBlankDeploymentLog(latestUvn, auditlog.Cancelled, false, "", "")
			dbClient.InsertValueInList(key, newCancelledLog)
		} else {
			latestLog.SetStatus(auditlog.Cancelled)
			dbClient.UpdateHeadInList(key, latestLog)
		}
	}
	w.WriteHeader(http.StatusOK)
	return
}

func InitStatusEndpoints(r *mux.Router) {
	r.HandleFunc("/status/{orgName}/{teamName}/pipeline/{pipelineName}/history/{count}", getPipelineUvns).Methods("GET")
	r.HandleFunc("/status/{orgName}/{teamName}/pipeline/{pipelineName}/step/{stepName}/{count}", getStepLogs).Methods("GET")
	r.HandleFunc("/status/{orgName}/{teamName}/pipeline/{pipelineName}/{pipelineUvn}", getPipelineStatus).Methods("GET")
	r.HandleFunc("/status/{orgName}/{teamName}/pipelineRun/{pipelineName}", cancelLatestPipeline).Methods("DELETE")
}
