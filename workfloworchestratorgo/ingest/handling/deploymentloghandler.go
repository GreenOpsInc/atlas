package handling

import (
	"github.com/greenopsinc/util/auditlog"
	cr "github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/event"
	"github.com/greenopsinc/workfloworchestrator/ingest/dbkey"
)

type DeploymentLogHandler interface {
	UpdateStepDeploymentLog(event event.Event, stepName string, argoApplicationName string, revisionHash string)
	InitializeNewStepLog(event event.Event, stepName string, pipelineUvn string, gitCommitVersion string)
	InitializeNewRemediationLog(event event.Event, stepName string, pipelineUvn string, resourceGVKList []*cr.ResourceGVK)
	MarkDeploymentSuccessful(event event.Event, stepName string)
	MarkStepSuccessful(event event.Event, stepName string)
	MarkStateRemediated(event event.Event, stepName string)
	MarkStateRemediationFailed(event event.Event, stepName string)
	MarkStepFailedWithFailedDeployment(event event.Event, stepName string)
	MarkStepFailedWithBrokenTest(event event.Event, stepName string, testName string, testLog string)
	MarkStepFailedWithProcessingError(event event.FailureEvent, stepName string, String error)
	AreParentStepsComplete(event event.Event, parentSteps []string) bool
	GetStepStatus(event event.Event) string
	MakeRollbackDeploymentLog(event event.Event, stepName string, rollbackLimit int, dryRun bool) string
	GetCurrentGitCommitHash(event event.Event, stepName string) string
	GetCurrentArgoRevisionHash(event event.Event, stepName string) string
	GetCurrentPipelineUvn(event event.Event, stepName string) string
	GetLastSuccessfulStepGitCommitHash(event event.Event, stepName string) string
	GetLastSuccessfulDeploymentGitCommitHash(event event.Event, stepName string) string
	GetLatestDeploymentLog(event event.Event, stepName string) *auditlog.DeploymentLog
}

type deploymentLogHandler struct {
	dbClient db.DbClient
}

func NewDeploymentLogHandler(dbClient db.DbClient) DeploymentLogHandler {
	return &deploymentLogHandler{dbClient: dbClient}
}

func (d *deploymentLogHandler) UpdateStepDeploymentLog(event event.Event, stepName string, argoApplicationName string, revisionHash string) {
	logKey := dbkey.MakeDbStepKey(event.GetOrgName(), event.GetTeamName(), event.GetPipelineName(), stepName)
	deploymentLog := d.dbClient.FetchLatestDeploymentLog(logKey)
	deploymentLog.SetArgoApplicationName(argoApplicationName)
	deploymentLog.SetArgoRevisionHash(revisionHash)
	d.dbClient.UpdateHeadInList(logKey, deploymentLog)
}

func (d *deploymentLogHandler) InitializeNewStepLog(event event.Event, stepName string, pipelineUvn string, gitCommitVersion string) {
	logKey := dbkey.MakeDbStepKey(event.GetOrgName(), event.GetTeamName(), event.GetPipelineName(), stepName)
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

func (d *deploymentLogHandler) InitializeNewRemediationLog(event event.Event, stepName string, pipelineUvn string, resourceGVKList []*cr.ResourceGVK) {
	logKey := dbkey.MakeDbStepKey(event.GetOrgName(), event.GetTeamName(), event.GetPipelineName(), stepName)
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

func (d *deploymentLogHandler) MarkDeploymentSuccessful(event event.Event, stepName string) {
	panic("implement me")
}

func (d *deploymentLogHandler) MarkStepSuccessful(event event.Event, stepName string) {
	panic("implement me")
}

func (d *deploymentLogHandler) MarkStateRemediated(event event.Event, stepName string) {
	panic("implement me")
}

func (d *deploymentLogHandler) MarkStateRemediationFailed(event event.Event, stepName string) {
	panic("implement me")
}

func (d *deploymentLogHandler) MarkStepFailedWithFailedDeployment(event event.Event, stepName string) {
	panic("implement me")
}

func (d *deploymentLogHandler) MarkStepFailedWithBrokenTest(event event.Event, stepName string, testName string, testLog string) {
	panic("implement me")
}

func (d *deploymentLogHandler) MarkStepFailedWithProcessingError(event interface{}, stepName string, String error) {
	panic("implement me")
}

func (d *deploymentLogHandler) AreParentStepsComplete(event event.Event, parentSteps []string) bool {
	panic("implement me")
}

func (d *deploymentLogHandler) GetStepStatus(event event.Event) string {
	panic("implement me")
}

func (d *deploymentLogHandler) MakeRollbackDeploymentLog(event event.Event, stepName string, rollbackLimit int, dryRun bool) string {
	panic("implement me")
}

func (d *deploymentLogHandler) GetCurrentGitCommitHash(event event.Event, stepName string) string {
	panic("implement me")
}

func (d *deploymentLogHandler) GetCurrentArgoRevisionHash(event event.Event, stepName string) string {
	panic("implement me")
}

func (d *deploymentLogHandler) GetCurrentPipelineUvn(event event.Event, stepName string) string {
	panic("implement me")
}

func (d *deploymentLogHandler) GetLastSuccessfulStepGitCommitHash(event event.Event, stepName string) string {
	panic("implement me")
}

func (d *deploymentLogHandler) GetLastSuccessfulDeploymentGitCommitHash(event event.Event, stepName string) string {
	panic("implement me")
}

func (d *deploymentLogHandler) GetLatestDeploymentLog(event event.Event, stepName string) *auditlog.DeploymentLog {
	panic("implement me")
}
