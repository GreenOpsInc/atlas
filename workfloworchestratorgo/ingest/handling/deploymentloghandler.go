package handling

import (
	"errors"
	"log"

	"github.com/greenopsinc/util/pipeline/data"

	"github.com/greenopsinc/util/auditlog"
	cr "github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/event"
	"github.com/greenopsinc/workfloworchestrator/ingest/dbkey"
)

type DeploymentLogHandler interface {
	UpdateStepDeploymentLog(e event.Event, stepName string, argoApplicationName string, revisionHash string)
	InitializeNewStepLog(e event.Event, stepName string, pipelineUvn string, gitCommitVersion string)
	InitializeNewRemediationLog(e event.Event, stepName string, pipelineUvn string, resourceGVKList []*cr.ResourceGVK)
	MarkDeploymentSuccessful(e event.Event, stepName string)
	MarkStepSuccessful(e event.Event, stepName string) error
	MarkStateRemediated(e event.Event, stepName string)
	MarkStateRemediationFailed(e event.Event, stepName string)
	MarkStepFailedWithFailedDeployment(e event.Event, stepName string)
	MarkStepFailedWithBrokenTest(e event.Event, stepName string, testName string, testLog string)
	MarkStepFailedWithProcessingError(e *event.FailureEvent, stepName string, error string)
	AreParentStepsComplete(e event.Event, parentSteps []string) bool
	GetStepStatus(e event.Event) string
	MakeRollbackDeploymentLog(e event.Event, stepName string, rollbackLimit int, dryRun bool) (string, error)
	GetCurrentGitCommitHash(e event.Event, stepName string) (string, error)
	GetCurrentArgoRevisionHash(e event.Event, stepName string) (string, error)
	GetCurrentPipelineUvn(e event.Event, stepName string) string
	GetLastSuccessfulStepGitCommitHash(e event.Event, stepName string) string
	GetLastSuccessfulDeploymentGitCommitHash(e event.Event, stepName string) string
	GetLatestDeploymentLog(e event.Event, stepName string) (*auditlog.DeploymentLog, error)
}

type deploymentLogHandler struct {
	dbClient db.DbClient
}

func NewDeploymentLogHandler(dbClient db.DbClient) DeploymentLogHandler {
	return &deploymentLogHandler{dbClient: dbClient}
}

func (d *deploymentLogHandler) UpdateStepDeploymentLog(e event.Event, stepName string, argoApplicationName string, revisionHash string) {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), stepName)
	deploymentLog := d.dbClient.FetchLatestDeploymentLog(logKey)
	deploymentLog.SetArgoApplicationName(argoApplicationName)
	deploymentLog.SetArgoRevisionHash(revisionHash)
	d.dbClient.UpdateHeadInList(logKey, deploymentLog)
}

func (d *deploymentLogHandler) InitializeNewStepLog(e event.Event, stepName string, pipelineUvn string, gitCommitVersion string) {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), stepName)
	newLog := &auditlog.DeploymentLog{
		PipelineUniqueVersionNumber: pipelineUvn,
		RollbackUniqueVersionNumber: "",
		UniqueVersionInstance:       0,
		Status:                      auditlog.Progressing,
		DeploymentComplete:          false,
		ArgoApplicationName:         "",
		ArgoRevisionHash:            "",
		GitCommitVersion:            gitCommitVersion,
		BrokenTest:                  "",
		BrokenTestLog:               "",
	}
	d.dbClient.InsertValueInList(logKey, newLog)
}

func (d *deploymentLogHandler) InitializeNewRemediationLog(e event.Event, stepName string, pipelineUvn string, resourceGVKList []*cr.ResourceGVK) {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), stepName)
	latestRemediationLog := d.dbClient.FetchLatestRemediationLog(logKey)
	if latestRemediationLog == nil || latestRemediationLog.GetPipelineUniqueVersionNumber() != pipelineUvn {
		latestRemediationLog = &auditlog.RemediationLog{
			PipelineUniqueVersionNumber: pipelineUvn,
			UniqueVersionInstance:       0,
			UnhealthyResources:          []string{},
			RemediationStatus:           auditlog.Progressing,
		}
	}

	var resourceNames []string
	for _, res := range resourceGVKList {
		resourceNames = append(resourceNames, res.ResourceName)
	}
	newLog := &auditlog.RemediationLog{
		PipelineUniqueVersionNumber: pipelineUvn,
		UniqueVersionInstance:       latestRemediationLog.GetUniqueVersionInstance() + 1,
		UnhealthyResources:          resourceNames,
		RemediationStatus:           auditlog.Progressing,
	}
	d.dbClient.InsertValueInList(logKey, newLog)
}

func (d *deploymentLogHandler) MarkDeploymentSuccessful(e event.Event, stepName string) {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), stepName)
	deploymentLog := d.dbClient.FetchLatestDeploymentLog(logKey)
	deploymentLog.SetDeploymentComplete(true)
	d.dbClient.UpdateHeadInList(logKey, deploymentLog)
}

func (d *deploymentLogHandler) MarkStepSuccessful(e event.Event, stepName string) error {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), stepName)
	deploymentLog := d.dbClient.FetchLatestDeploymentLog(logKey)

	// this check is largely redundant, should never be the case
	if deploymentLog.GetBrokenTest() != "" {
		return errors.New("this step has test failures, should not be marked successful")
	}
	deploymentLog.SetStatus(auditlog.Success)
	d.dbClient.UpdateHeadInList(logKey, deploymentLog)
	return nil
}

func (d *deploymentLogHandler) MarkStateRemediated(e event.Event, stepName string) {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), stepName)
	remediationLog := d.dbClient.FetchLatestRemediationLog(logKey)
	if remediationLog == nil {
		log.Println("no remediation log present")
		return
	}

	remediationLog.SetStatus(auditlog.Success)
	d.dbClient.UpdateHeadInList(logKey, remediationLog)
}

func (d *deploymentLogHandler) MarkStateRemediationFailed(e event.Event, stepName string) {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), stepName)
	remediationLog := d.dbClient.FetchLatestRemediationLog(logKey)
	if remediationLog == nil {
		log.Println("no remediation log present")
		return
	}

	remediationLog.SetStatus(auditlog.Failure)
	d.dbClient.UpdateHeadInList(logKey, remediationLog)
}

func (d *deploymentLogHandler) MarkStepFailedWithFailedDeployment(e event.Event, stepName string) {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), stepName)
	deploymentLog := d.dbClient.FetchLatestDeploymentLog(logKey)
	deploymentLog.SetDeploymentComplete(false)
	deploymentLog.SetStatus(auditlog.Failure)
	d.dbClient.UpdateHeadInList(logKey, deploymentLog)
}

func (d *deploymentLogHandler) MarkStepFailedWithBrokenTest(e event.Event, stepName string, testName string, testLog string) {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), stepName)
	deploymentLog := d.dbClient.FetchLatestDeploymentLog(logKey)
	deploymentLog.SetBrokenTest(testName)
	deploymentLog.SetBrokenTestLog(testLog)
	deploymentLog.SetStatus(auditlog.Failure)
	d.dbClient.UpdateHeadInList(logKey, deploymentLog)
}

func (d *deploymentLogHandler) MarkStepFailedWithProcessingError(e *event.FailureEvent, stepName string, error string) {
	log.Println("marking step failed with processing error")
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), stepName)
	log := d.dbClient.FetchLatestLog(logKey)
	if log == nil {
		d.markPipelineFailedWithProcessingError(e, error)
		return
	} else {
		_, ok := log.(*auditlog.DeploymentLog)
		if ok {
			log.SetBrokenTest("Processing Error")
			log.SetBrokenTestLog(error)
		}
	}
	log.SetStatus(auditlog.Failure)
	d.dbClient.UpdateHeadInList(logKey, log)
}

func (d *deploymentLogHandler) AreParentStepsComplete(e event.Event, parentSteps []string) bool {
	for _, parentStepName := range parentSteps {
		if parentStepName == data.RootStepName {
			continue
		}
		logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), parentStepName)
		deploymentLog := d.dbClient.FetchLatestDeploymentLog(logKey)
		if deploymentLog.GetUniqueVersionInstance() != 0 || deploymentLog.GetStatus() != auditlog.Success {
			return false
		}
	}
	return true
}

func (d *deploymentLogHandler) GetStepStatus(e event.Event) string {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), e.GetStepName())
	deploymentLog := d.dbClient.FetchLatestDeploymentLog(logKey)
	return string(deploymentLog.GetStatus())
}

// MakeRollbackDeploymentLog
// returning null means a failure occurred, empty string means no match exists
func (d *deploymentLogHandler) MakeRollbackDeploymentLog(e event.Event, stepName string, rollbackLimit int, dryRun bool) (string, error) {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), e.GetStepName())
	logIncrement := 0
	logList := d.dbClient.FetchLogList(logKey, logIncrement)
	if logList == nil || len(logList) == 0 {
		return "", nil
	}

	currentLog := d.dbClient.FetchLatestDeploymentLog(logKey)
	if currentLog == nil {
		return "", errors.New("a rollback was triggered, but no previous logs could be found for the step")
	}
	if currentLog.GetUniqueVersionInstance() >= rollbackLimit {
		log.Println("met the limit for rollbacks for this step, so returning an invalid commit hash to stop further processing")
		return "", nil
	}

	// this means that there was probably an error during the execution of the step, and that the log was added but the re-triggering process was not completed
	if currentLog.GetStatus() == auditlog.Progressing && currentLog.GetUniqueVersionInstance() > 0 {
		return currentLog.GetGitCommitVersion(), nil
	}

	idx := 0
	// if a specific rollback UVN has already been tried but has failed, we want to skip to the first instance of that UVN and search before then
	if currentLog.GetUniqueVersionInstance() > 0 {
		for idx < len(logList) {
			if currentLog.GetRollbackUniqueVersionNumber() == logList[idx].GetPipelineUniqueVersionNumber() &&
				logList[idx].GetUniqueVersionInstance() == 0 {
				idx++
				break
			}

			idx++
			if idx == len(logList) {
				logIncrement++
				logList = d.dbClient.FetchLogList(logKey, logIncrement)
				idx = 0
			}
		}
	} else if currentLog.GetStatus() == auditlog.Success {
		for idx < len(logList) {
			_, ok := logList[idx].(*auditlog.DeploymentLog)
			if !ok {
				idx++
				continue
			}
			if currentLog.GetPipelineUniqueVersionNumber() != logList[idx].GetPipelineUniqueVersionNumber() {
				break
			}

			idx++
			if idx == len(logList) {
				logIncrement++
				logList = d.dbClient.FetchLogList(logKey, logIncrement)
				idx = 0
			}
		}
	}

	for idx < len(logList) {
		deploymentLog, ok := logList[idx].(*auditlog.DeploymentLog)
		if !ok {
			idx++
			continue
		}
		if deploymentLog.GetStatus() == auditlog.Success && deploymentLog.GetUniqueVersionInstance() == 0 {
			gitCommitVersion := deploymentLog.GetGitCommitVersion()
			logKey = dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), e.GetStepName())
			newLog := &auditlog.DeploymentLog{
				PipelineUniqueVersionNumber: currentLog.GetPipelineUniqueVersionNumber(),
				RollbackUniqueVersionNumber: logList[idx].GetPipelineUniqueVersionNumber(),
				UniqueVersionInstance:       currentLog.GetUniqueVersionInstance() + 1,
				Status:                      auditlog.Progressing,
				DeploymentComplete:          false,
				ArgoApplicationName:         deploymentLog.GetArgoApplicationName(),
				ArgoRevisionHash:            deploymentLog.GetArgoRevisionHash(),
				GitCommitVersion:            gitCommitVersion,
				BrokenTest:                  "",
				BrokenTestLog:               "",
			}

			if !dryRun {
				d.dbClient.InsertValueInList(logKey, newLog)
			}
			return gitCommitVersion, nil
		}

		idx++
		if idx == len(logList) {
			logIncrement++
			logList = d.dbClient.FetchLogList(logKey, logIncrement)
			idx = 0
		}
	}
	return "", nil
}

func (d *deploymentLogHandler) GetCurrentGitCommitHash(e event.Event, stepName string) (string, error) {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), e.GetStepName())
	currentDeploymentLog := d.dbClient.FetchLatestDeploymentLog(logKey)
	if currentDeploymentLog == nil {
		return "", errors.New("no deployment log found for this key, no commit hash will be found")
	}
	return currentDeploymentLog.GetGitCommitVersion(), nil
}

func (d *deploymentLogHandler) GetCurrentArgoRevisionHash(e event.Event, stepName string) (string, error) {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), e.GetStepName())
	currentDeploymentLog := d.dbClient.FetchLatestDeploymentLog(logKey)
	if currentDeploymentLog == nil {
		return "", errors.New("no deployment log found for this key, no commit hash will be found")
	}
	return currentDeploymentLog.GetGitCommitVersion(), nil
}

func (d *deploymentLogHandler) GetCurrentPipelineUvn(e event.Event, stepName string) string {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), e.GetStepName())
	currentDeploymentLog := d.dbClient.FetchLatestDeploymentLog(logKey)
	if currentDeploymentLog == nil {
		return ""
	}
	return currentDeploymentLog.GetPipelineUniqueVersionNumber()
}

func (d *deploymentLogHandler) GetLastSuccessfulStepGitCommitHash(e event.Event, stepName string) string {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), e.GetStepName())
	logIncrement := 0
	logList := d.dbClient.FetchLogList(logKey, logIncrement)
	if logList == nil || len(logList) == 0 {
		return ""
	}

	idx := 0
	for idx < len(logList) {
		deploymentLog, ok := logList[idx].(*auditlog.DeploymentLog)
		if !ok {
			idx++
			continue
		}
		if deploymentLog.GetStatus() == auditlog.Success {
			return deploymentLog.GetGitCommitVersion()
		}

		idx++
		if idx == len(logList) {
			logIncrement++
			logList = d.dbClient.FetchLogList(logKey, logIncrement)
			idx = 0
		}
	}
	return ""
}

func (d *deploymentLogHandler) GetLastSuccessfulDeploymentGitCommitHash(e event.Event, stepName string) string {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), e.GetStepName())
	logIncrement := 0
	logList := d.dbClient.FetchLogList(logKey, logIncrement)
	if logList == nil || len(logList) == 0 {
		return ""
	}

	idx := 0
	for idx < len(logList) {
		deploymentLog, ok := logList[idx].(*auditlog.DeploymentLog)
		if !ok {
			idx++
			continue
		}
		if deploymentLog.IsDeploymentComplete() {
			return deploymentLog.GetGitCommitVersion()
		}

		idx++
		if idx == len(logList) {
			logIncrement++
			logList = d.dbClient.FetchLogList(logKey, logIncrement)
			idx = 0
		}
	}
	return ""
}

func (d *deploymentLogHandler) GetLatestDeploymentLog(e event.Event, stepName string) (*auditlog.DeploymentLog, error) {
	logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), stepName)
	deploymentLog, ok := d.dbClient.FetchLatestDeploymentLog(logKey).(*auditlog.DeploymentLog)
	if !ok {
		return nil, errors.New("og type is not DeploymentLog")
	}
	return deploymentLog, nil
}

// when the deployment log for a step doesn't exist, just mark the entire pipeline ambiguously and allow the user to determine the origin
func (d *deploymentLogHandler) markPipelineFailedWithProcessingError(e *event.FailureEvent, error string) {
	log.Println("no step found for the error, adding in to the pipeline metadata")
	key := dbkey.MakeDbPipelineInfoKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName())
	pipelineInfo := d.dbClient.FetchLatestPipelineInfo(key)

	if e.StatusCode == event.PipelineTriggerEventName {
		pipelineInfo = auditlog.PipelineInfo{
			PipelineUvn: e.GetUVN(),
			Errors:      []string{error},
			StepList:    []string{},
		}
		d.dbClient.InsertValueInList(key, pipelineInfo)
	} else {
		pipelineInfo.Errors = append(pipelineInfo.Errors, error)
		d.dbClient.UpdateHeadInList(key, pipelineInfo)
	}
}
