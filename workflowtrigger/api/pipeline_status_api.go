package api

import (
	"github.com/gorilla/mux"
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

func getPipelineStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	pipelineUvn := vars[pipelineUvnField]

	status := pipelinestatus.New()
	steps := dbClient.FetchStringList(db.MakeDbListOfStepsKey(orgName, teamName, pipelineName))
	if len(steps) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(""))
		return
	}
	for _, step := range steps {
		//Get pipeline UVN if not specified
		logKey := db.MakeDbStepKey(orgName, teamName, pipelineName, step)
		var log auditlog.Log
		if pipelineUvn == "LATEST" {
			log = dbClient.FetchLatestLog(logKey)
			if log == nil {
				http.Error(w, "No deployment log exists", http.StatusBadRequest)
				return
			}
			pipelineUvn = log.GetPipelineUniqueVersionNumber()
		}

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
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(serializer.Serialize(status)))
}

func cancelLatestPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	teamName := vars[teamNameField]
	pipelineName := vars[pipelineNameField]
	latestUvn := ""
	steps := dbClient.FetchStringList(db.MakeDbListOfStepsKey(orgName, teamName, pipelineName))
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
	r.HandleFunc("/status/{orgName}/{teamName}/pipeline/{pipelineName}/step/{stepName}/{count}", getStepLogs).Methods("GET")
	r.HandleFunc("/status/{orgName}/{teamName}/pipeline/{pipelineName}/{pipelineUvn}", getPipelineStatus).Methods("GET")
	r.HandleFunc("/status/{orgName}/{teamName}/pipelineRun/{pipelineName}", cancelLatestPipeline).Methods("DELETE")
}