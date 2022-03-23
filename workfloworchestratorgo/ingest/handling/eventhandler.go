package handling

import (
	"errors"
	"log"

	"github.com/greenopsinc/util/array"
	"github.com/greenopsinc/util/auditlog"
	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/event"
	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/pipeline/data"
	"github.com/greenopsinc/util/team"
	"github.com/greenopsinc/workfloworchestrator/ingest/apiclient/reposerver"
	"github.com/greenopsinc/workfloworchestrator/ingest/dbkey"
	"gopkg.in/yaml.v2"
)

type WatchKeyType string

const (
	WatchArgoApplicationKey WatchKeyType = "WatchArgoApplicationKey"
	WatchTestKey            WatchKeyType = "WatchTestKey"
	WatchArgoWorkflowKey    WatchKeyType = "ArgoWorkflowTask"
	PipelineFileName                     = "pipeline.yaml"
)

type EventHandler interface {
	HandleEvent(e event.Event) error
}

type eventHandler struct {
	repoManagerApi       reposerver.RepoManagerAPI
	dbClient             db.DbClient
	deploymentHandler    DeploymentHandler
	testHandler          TestHandler
	deploymentLogHandler DeploymentLogHandler
	// TODO: implement kubernetes work queue
	//wqClient             WQClient
}

func NewEventHandler(
	repoManagerApi reposerver.RepoManagerAPI,
	dbClient db.DbClient,
	deploymentHandler DeploymentHandler,
	testHandler TestHandler,
	deploymentLogHandler DeploymentLogHandler,
	//mqClient MQClient,
) EventHandler {
	return &eventHandler{
		repoManagerApi:       repoManagerApi,
		dbClient:             dbClient,
		deploymentHandler:    deploymentHandler,
		testHandler:          testHandler,
		deploymentLogHandler: deploymentLogHandler,
		//mqClient:             mqClient,
	}
}

func (eh *eventHandler) HandleEvent(e event.Event) error {
	gitCommit := ""
	teamSchema := eh.fetchTeamSchema(e)
	if teamSchema == nil {
		return errors.New("the team doesn't exist")
	}

	// this checks whether the event was from a previous pipeline run
	// if it is, it will be ignored.
	isStaleEvent, err := eh.isStaleEvent(e)
	if err != nil {
		return err
	}
	if isStaleEvent {
		log.Printf("event from pipeline %s is stale, ignoring...", e.GetUVN())
	}

	if e.GetStepName() != data.RootStepName {
		gitCommitRes, err := eh.deploymentLogHandler.GetCurrentGitCommitHash(e, e.GetStepName())
		if err != nil {
			return err
		}
		gitCommit = gitCommitRes
	}

	tempGitRepoSchema := teamSchema.GetPipelineSchema(e.GetPipelineName()).GetGitRepoSchema()
	gitRepoSchemaInfo := &git.GitRepoSchemaInfo{
		GitRepoUrl: tempGitRepoSchema.GetGitRepo(),
		PathToRoot: tempGitRepoSchema.GetPathToRoot(),
	}
	pipelineData, err := eh.fetchPipelineData(e, gitRepoSchemaInfo, gitCommit)
	if err != nil {
		return err
	}
	if pipelineData == nil {
		return errors.New("the pipeline doesn't exist")
	}

	switch ev := e.(type) {
	case *event.PipelineTriggerEvent:
		log.Printf("Handling event of type PipelineTriggerEvent")
		gitCommit = ev.RevisionHash
		return eh.handlePipelineTriggerEvent(pipelineData, gitRepoSchemaInfo, ev)
	case *event.ApplicationInfraCompletionEvent:
		log.Printf("Handling event of type ApplicationInfraCompletionEvent")
		return eh.handleApplicationInfraCompletion(pipelineData, gitRepoSchemaInfo, ev)
	case *event.ApplicationInfraTriggerEvent:
		log.Printf("Handling event of type ApplicationInfraTriggerEvent")
		return eh.handleApplicationInfraTrigger(pipelineData, teamSchema, gitRepoSchemaInfo, ev)
	case *event.ClientCompletionEvent:
		log.Printf("Handling event of type ClientCompletionEvent")
		return eh.handleClientCompletionEvent(pipelineData, gitRepoSchemaInfo, ev)
	case *event.FailureEvent:
		log.Printf("Handling event of type FailureEvent")
		eh.handleFailureEvent(ev)
		return nil
	case *event.TestCompletionEvent:
		log.Printf("Handling event of type TestCompletionEvent")
		return eh.handleTestCompletion(pipelineData, gitRepoSchemaInfo, ev)
	case *event.TriggerStepEvent:
		log.Printf("Handling event of type TriggerStepEvent")
		gitCommit = ev.GitCommitHash
		return eh.handleTriggerStep(pipelineData, gitRepoSchemaInfo, ev)
	default:
		return errors.New("wrong event object provided")
	}
}

func (eh *eventHandler) isStaleEvent(e event.Event) (bool, error) {
	if _, ok := e.(*event.PipelineTriggerEvent); ok {
		return false, nil
	}

	deploymentLog, err := eh.deploymentLogHandler.GetLatestDeploymentLog(e, e.GetStepName())
	if err != nil {
		return false, err
	}

	_, isTestCompletionEvent := e.(*event.TestCompletionEvent)
	_, isApplicationInfraTriggerEvent := e.(*event.ApplicationInfraTriggerEvent)
	_, isApplicationInfraCompletionEvent := e.(*event.ApplicationInfraCompletionEvent)
	_, isFailureEvent := e.(*event.FailureEvent)
	_, isTriggerStepEvent := e.(*event.TriggerStepEvent)
	if deploymentLog != nil &&
		(deploymentLog.GetStatus() == auditlog.Success || deploymentLog.GetStatus() == auditlog.Failure || deploymentLog.GetStatus() == auditlog.Cancelled) &&
		(isTestCompletionEvent || isApplicationInfraTriggerEvent || isApplicationInfraCompletionEvent || isFailureEvent) {
		return true, nil
	} else if (deploymentLog != nil && deploymentLog.GetStatus() == auditlog.Progressing) && isTriggerStepEvent {
		return true, nil
	}

	currentUvn := eh.deploymentLogHandler.GetCurrentPipelineUvn(e, e.GetStepName())
	if e.GetUVN() != currentUvn && isTriggerStepEvent {
		return false, nil
	}
	return currentUvn != "" && e.GetUVN() != currentUvn, nil
}

func (eh *eventHandler) handlePipelineTriggerEvent(pipelineData *data.PipelineData, gitRepoSchemaInfo *git.GitRepoSchemaInfo, e *event.PipelineTriggerEvent) error {
	pipelineInfoKey := dbkey.MakeDbPipelineInfoKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName())
	pipelineInfo := eh.dbClient.FetchLatestPipelineInfo(pipelineInfoKey)
	listOfSteps := pipelineInfo.StepList

	if listOfSteps != nil && len(listOfSteps) > 0 {
		latestUVN := pipelineInfo.PipelineUvn
		progressing := false
		logKey := dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), listOfSteps[0])
		deploymentLog := eh.dbClient.FetchLatestDeploymentLog(logKey)
		if deploymentLog != nil || deploymentLog.GetPipelineUniqueVersionNumber() != latestUVN {
			if len(pipelineInfo.Errors) == 0 {
				// if there is a list but the deployment log is null it means the first step has not been triggered yet
				eh.queueNewPipelineRun(e, latestUVN)
			} else {
				// errors combined with no logs for the first step means there was an error and the pipeline is complete
				if err := eh.initializeNewPipelineRun(e, gitRepoSchemaInfo, pipelineData); err != nil {
					return err
				}
			}
			return nil
		}
		currentPipelineData, err := eh.fetchPipelineData(e, gitRepoSchemaInfo, deploymentLog.GetGitCommitVersion())
		if err != nil {
			return err
		}

		// there are two options for starting steps:
		// 1. either the pipeline was run normally, and the starting steps are the root steps
		// 2. a sub-pipeline was run, and there is an arbitrary starting step
		orderedStepsBfs := make([]string, 0)
		rootSteps := currentPipelineData.StepChildren[data.RootStepName]
		for _, stepName := range rootSteps {
			if array.ContainElement(pipelineInfo.StepList, stepName) {
				orderedStepsBfs = append(orderedStepsBfs, stepName)
			}
		}
		if len(orderedStepsBfs) == 0 {
			orderedStepsBfs = append(orderedStepsBfs, pipelineInfo.StepList[0])
		}

		// iterates through the logs to verify completion
		idx := 0
		for idx < len(orderedStepsBfs) {
			logKey = dbkey.MakeDbStepKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName(), orderedStepsBfs[idx])
			deploymentLog = eh.dbClient.FetchLatestDeploymentLog(logKey)
			if deploymentLog == nil || deploymentLog.GetPipelineUniqueVersionNumber() != latestUVN {
				progressing = true
				break
			}
			if deploymentLog.GetPipelineUniqueVersionNumber() == latestUVN && deploymentLog.GetStatus() == auditlog.Progressing {
				progressing = true
				break
			} else if deploymentLog.GetPipelineUniqueVersionNumber() == latestUVN && deploymentLog.GetStatus() == auditlog.Failure {
				stepData := currentPipelineData.GetStep(orderedStepsBfs[idx])
				// if the step has rollbacks enabled but has failed & not hit its rollback limit, the pipeline is still progressing
				// the second part of the expression ensures that there is not a valid version to rollback to (if there isn't, the rollback limit may not have been hit)
				rollbackDeploymentLog, err := eh.deploymentLogHandler.MakeRollbackDeploymentLog(e, stepData.Name, stepData.RollbackLimit, true)
				if err != nil {
					return err
				}
				if stepData.RollbackLimit > deploymentLog.GetUniqueVersionInstance() &&
					rollbackDeploymentLog == "" {
					progressing = true
					break
				}
			} else if deploymentLog.GetPipelineUniqueVersionNumber() == latestUVN && deploymentLog.GetStatus() == auditlog.Success {
				for _, el := range currentPipelineData.StepChildren[orderedStepsBfs[idx]] {
					if array.ContainElement(pipelineInfo.StepList, el) {
						orderedStepsBfs = append(orderedStepsBfs, el)
					}
				}
			}
			idx++
		}
		if progressing {
			eh.queueNewPipelineRun(e, latestUVN)
			return nil
		}
	}
	return eh.initializeNewPipelineRun(e, gitRepoSchemaInfo, pipelineData)
}

func (eh *eventHandler) queueNewPipelineRun(e *event.PipelineTriggerEvent, latestUVN string) {
	// TODO: send message via wq client
	// eh.wqClient.SendMessage(e)
	log.Printf("Pipeline %s in progress, queueing up pipeline %s", latestUVN, e.GetUVN())
}

func (eh *eventHandler) initializeNewPipelineRun(e *event.PipelineTriggerEvent, gitRepoSchemaInfo *git.GitRepoSchemaInfo, pipelineData *data.PipelineData) error {
	pipelineInfoKey := dbkey.MakeDbPipelineInfoKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName())
	log.Printf("starting new run for pipeline %s", e.GetPipelineName())
	if e.GetStepName() == data.RootStepName {
		eh.dbClient.InsertValueInList(pipelineInfoKey, &auditlog.PipelineInfo{
			PipelineUvn: e.GetUVN(),
			Errors:      []string{},
			StepList:    pipelineData.GetAllStepsOrdered(),
		})
		if err := eh.triggerNextSteps(pipelineData, data.CreateRootStep(), gitRepoSchemaInfo, e); err != nil {
			return err
		}
	} else {
		eh.dbClient.InsertValueInList(pipelineInfoKey, &auditlog.PipelineInfo{
			PipelineUvn: e.GetUVN(),
			Errors:      []string{},
			StepList:    []string{e.GetStepName()},
		})
		// TODO: send message via wq client
		//eh.wqClient.SendMessage(&event.TriggerStepEvent{
		//	OrgName:       e.GetOrgName(),
		//	TeamName:      e.GetTeamName(),
		//	PipelineName:  e.GetPipelineName(),
		//	StepName:      e.GetStepName(),
		//	UVN:           e.GetUVN(),
		//	GitCommitHash: e.RevisionHash,
		//	Rollback:      false,
		//})
	}
	return nil
}

func (eh *eventHandler) handleClientCompletionEvent(pipelineData *data.PipelineData, gitRepoSchemaInfo *git.GitRepoSchemaInfo, e *event.ClientCompletionEvent) error {
	if auditlog.LogStatus(e.HealthStatus) == auditlog.Progressing {
		return nil
	}

	step := pipelineData.GetStep(e.StepName)
	logKey := dbkey.MakeDbStepKey(e.OrgName, e.TeamName, e.PipelineName, e.StepName)
	latestLog := eh.dbClient.FetchLatestLog(logKey)
	if latestLog != nil && latestLog.GetStatus() == auditlog.Cancelled {
		return nil
	}

	deploymentLog := eh.dbClient.FetchLatestDeploymentLog(logKey)
	if deploymentLog == nil {
		return nil
	}
	if deploymentLog.GetStatus() == auditlog.Failure || deploymentLog.GetStatus() == auditlog.Cancelled {
		return nil
	} else if deploymentLog.GetStatus() == auditlog.Success {
		// if the deployment was successful (rollback or otherwise), these client completion events are for state remediation
		if step.RemediationLimit == 0 {
			return nil
		}
		if e.SyncStatus == event.OutOfSyncStatus && e.HealthStatus == event.MissingStatus &&
			eh.deploymentHandler.RollbackInPipelineExists(e, pipelineData, step.Name) {
			return nil
		}

		remediationLog := eh.dbClient.FetchLatestRemediationLog(logKey)
		if remediationLog != nil && remediationLog.GetStatus() == auditlog.Cancelled {
			return nil
		}

		// TODO: right now, remediation is only based on health. We should be adding in pruning based on OutOfSync statuses as well.
		// TODO: right now the events are being sent but are just being ignored.
		if e.HealthStatus == event.DegradedStatus || e.HealthStatus == event.UnknownStatus {
			//                    || e.HealthStatus == event.MissingStatus {
			if remediationLog != nil {
				if remediationLog.GetStatus() == auditlog.Progressing {
					eh.deploymentLogHandler.MarkStateRemediationFailed(e, step.Name)
				}
				if remediationLog.GetUniqueVersionInstance() == step.RemediationLimit {
					// reached remediation limit
					if step.RollbackLimit > 0 {
						if err := eh.rollback(e); err != nil {
							return err
						}
					}
					return nil
				}
			}

			resourceGVKList := []*clientrequest.ResourceGVK{}
			for _, res := range e.ResourceStatuses {
				resourceGVKList = append(resourceGVKList, &clientrequest.ResourceGVK{
					ResourceName:      res.ResourceName,
					ResourceNamespace: res.ResourceNamespace,
					Group:             res.Group,
					Version:           res.Version,
					Kind:              res.Kind,
				})
			}
			eh.deploymentLogHandler.InitializeNewRemediationLog(e, step.Name, deploymentLog.GetPipelineUniqueVersionNumber(), resourceGVKList)
			if err := eh.deploymentHandler.TriggerStateRemediation(e, gitRepoSchemaInfo, step, deploymentLog.GetArgoApplicationName(), deploymentLog.GetArgoRevisionHash(), resourceGVKList); err != nil {
				return err
			}
		} else if e.HealthStatus == event.HealthyStatus {
			eh.deploymentLogHandler.MarkStateRemediated(e, step.Name)
		} else {
			eh.deploymentLogHandler.UpdateStepDeploymentLog(e, e.StepName, e.ArgoName, e.RevisionHash)
			if e.HealthStatus == event.DegradedStatus || e.HealthStatus == event.UnknownStatus {
				eh.deploymentLogHandler.MarkStepFailedWithFailedDeployment(e, e.StepName)
				if step.RollbackLimit > 0 {
					if err := eh.rollback(e); err != nil {
						return err
					}
				}
				return nil
			}
		}

		eh.deploymentLogHandler.MarkDeploymentSuccessful(e, e.StepName)
		var afterTestsExist bool
		for _, t := range step.Tests {
			if !t.ShouldExecuteBefore() {
				afterTestsExist = true
			}
		}
		if afterTestsExist {
			commit, err := eh.deploymentLogHandler.GetCurrentGitCommitHash(e, step.Name)
			if err != nil {
				return err
			}
			if err = eh.testHandler.TriggerTest(gitRepoSchemaInfo, step, false, commit, e); err != nil {
				return err
			}
		} else {
			if err := eh.triggerNextSteps(pipelineData, step, gitRepoSchemaInfo, e); err != nil {
				return err
			}
		}
	}
	return nil
}

func (eh *eventHandler) handleTestCompletion(pipelineData *data.PipelineData, gitRepoSchemaInfo *git.GitRepoSchemaInfo, e *event.TestCompletionEvent) error {
	step := pipelineData.GetStep(e.StepName)
	if !e.Successful {
		eh.deploymentLogHandler.MarkStepFailedWithBrokenTest(e, e.StepName, eh.getTestNameFromNumber(step, e.TestNumber), e.Log)
		if step.RollbackLimit > 0 {
			if err := eh.rollback(e); err != nil {
				return err
			}
		}
		return nil
	}

	completedTestNumber := e.TestNumber
	if completedTestNumber < 0 || len(step.Tests) <= completedTestNumber {
		log.Println("malformed test key or tests have changed, this event will be ignored")
		return nil
	}

	completedTest := step.Tests[completedTestNumber]
	var tests []*data.TestData
	for _, t := range step.Tests {
		if t.ShouldExecuteBefore() == completedTest.ShouldExecuteBefore() {
			tests = append(tests, &t)
		}
	}

	if completedTest.ShouldExecuteBefore() && completedTestNumber == len(tests)-1 {
		eh.triggerAppInfraDeploy(step.Name, e)
	} else if !completedTest.ShouldExecuteBefore() && completedTestNumber == len(step.Tests)-1 {
		// if there are before tasks, the numbering for "after" tasks won't be exact
		// the completed test number may go past the number of after tests
		if err := eh.triggerNextSteps(pipelineData, step, gitRepoSchemaInfo, e); err != nil {
			return err
		}
	} else if (completedTest.ShouldExecuteBefore() && completedTestNumber < len(tests)) ||
		(!completedTest.ShouldExecuteBefore() && completedTestNumber < len(step.Tests)) {
		hash, err := eh.deploymentLogHandler.GetCurrentGitCommitHash(e, step.Name)
		if err != nil {
			return err
		}
		if err = eh.testHandler.CreateAndRunTest(
			step.ClusterName,
			step,
			gitRepoSchemaInfo,
			step.Tests[completedTestNumber+1],
			completedTestNumber+1,
			hash,
			e,
		); err != nil {
			return err
		}
	} else {
		// this case should never be happening...log and see what the edge case is
		log.Printf("EDGE CASE: %s, %s", completedTest.ShouldExecuteBefore(), completedTestNumber == len(step.Tests)-1)
	}
	return nil
}

func (eh *eventHandler) handleApplicationInfraTrigger(pipelineData *data.PipelineData, teamSchema *team.TeamSchema, gitRepoSchemaInfo *git.GitRepoSchemaInfo, e *event.ApplicationInfraTriggerEvent) error {
	// right now it is assumed that the step names do not change
	stepData := pipelineData.GetStep(e.StepName)
	if stepData.OtherDeploymentsPath == "" {
		// TODO: send message via wq client
		//eh.wqClient.SendMessage(&event.ApplicationInfraCompletionEvent{
		//	OrgName:      e.GetOrgName(),
		//	TeamName:     e.GetTeamName(),
		//	PipelineName: e.GetPipelineName(),
		//	StepName:     e.GetStepName(),
		//	UVN:          e.GetUVN(),
		//	Success:      true,
		//})
		return nil
	}

	var oldGitCommitHash string
	if stepData.RollbackLimit > 0 {
		oldGitCommitHash = eh.deploymentLogHandler.GetLastSuccessfulStepGitCommitHash(e, e.StepName)
	} else {
		// TODO: add deploymentInfraSuccessful variable in deployment log, and replace method below with one that checks for the deployment infra deploying successfully
		oldGitCommitHash = eh.deploymentLogHandler.GetLastSuccessfulDeploymentGitCommitHash(e, e.StepName)
	}

	if oldGitCommitHash != "" {
		oldPipelineData, err := eh.fetchPipelineData(e, gitRepoSchemaInfo, oldGitCommitHash)
		if err != nil {
			return err
		}
		oldStepData := oldPipelineData.GetStep(e.StepName)
		// TODO: two separate events cannot be sent for deleting then deploying.
		// 		the deployment handler has to version application infrastructure with Kubernetes labels
		if err = eh.deploymentHandler.DeleteApplicationInfrastructure(e, gitRepoSchemaInfo, oldStepData, oldGitCommitHash); err != nil {
			return err
		}
	}
	hash, err := eh.deploymentLogHandler.GetCurrentGitCommitHash(e, e.StepName)
	if err != nil {
		return err
	}
	return eh.deploymentHandler.DeleteApplicationInfrastructure(e, gitRepoSchemaInfo, pipelineData.GetStep(e.StepName), hash)
}

func (eh *eventHandler) handleApplicationInfraCompletion(pipelineData *data.PipelineData, gitRepoSchemaInfo *git.GitRepoSchemaInfo, e *event.ApplicationInfraCompletionEvent) error {
	stepData := pipelineData.GetStep(e.StepName)
	if !e.Success && stepData.RollbackLimit > 0 {
		if err := eh.rollback(e); err != nil {
			return err
		}
	}

	logKey := dbkey.MakeDbStepKey(e.OrgName, e.TeamName, e.PipelineName, stepData.Name)
	deploymentLog := eh.dbClient.FetchLatestDeploymentLog(logKey)

	if (stepData.ArgoApplicationPath != "" || stepData.ArgoApplication != "") && deploymentLog.GetUniqueVersionInstance() > 0 {
		if err := eh.deploymentHandler.RollbackArgoApplication(e, gitRepoSchemaInfo, stepData, deploymentLog.GetArgoApplicationName(), deploymentLog.GetArgoRevisionHash()); err != nil {
			return err
		}
		return nil
	} else if stepData.ArgoApplicationPath != "" || stepData.ArgoApplication != "" {
		argoRevisionHash, err := eh.deploymentLogHandler.GetCurrentArgoRevisionHash(e, stepData.Name)
		if err != nil {
			return err
		}
		hash, err := eh.deploymentLogHandler.GetCurrentGitCommitHash(e, stepData.Name)
		return eh.deploymentHandler.DeployArgoApplication(e, gitRepoSchemaInfo, pipelineData, stepData.Name, argoRevisionHash, hash)
	}
	return nil
}

func (eh *eventHandler) handleTriggerStep(pipelineData *data.PipelineData, gitRepoSchemaInfo *git.GitRepoSchemaInfo, e *event.TriggerStepEvent) error {
	gitCommit := e.GitCommitHash

	// first cancel any pending remediation steps
	logKey := dbkey.MakeDbStepKey(e.OrgName, e.TeamName, e.PipelineName, e.StepName)
	latestLog := eh.dbClient.FetchLatestLog(logKey)
	_, isRemediationLog := latestLog.(*auditlog.RemediationLog)
	if isRemediationLog && latestLog.GetStatus() == auditlog.Progressing {
		latestLog.SetStatus(auditlog.Cancelled)
		eh.dbClient.UpdateHeadInList(logKey, latestLog)
	}

	if e.Rollback {
		log.Println("handling rollback trigger")
		gitCommitRes, err := eh.deploymentLogHandler.MakeRollbackDeploymentLog(e, e.StepName, pipelineData.GetStep(e.StepName).RollbackLimit, false)
		if err != nil {
			return err
		}
		if gitCommitRes == "" {
			log.Println("could not find stable deployment to rollback to")
			// means there is no stable version that can be found
			return nil
		}
		gitCommit = gitCommitRes
		pipelineData, err = eh.fetchPipelineData(e, gitRepoSchemaInfo, gitCommit)
		if err != nil {
			return err
		}
	} else {
		eh.deploymentLogHandler.InitializeNewStepLog(e, e.StepName, e.UVN, gitCommit)
	}

	stepData := pipelineData.GetStep(e.StepName)
	var beforeTestsExist bool
	for _, t := range stepData.Tests {
		if t.ShouldExecuteBefore() {
			beforeTestsExist = true
		}
	}
	if beforeTestsExist {
		return eh.testHandler.TriggerTest(gitRepoSchemaInfo, stepData, true, gitCommit, e)
	}

	if stepData.OtherDeploymentsPath != "" || stepData.ArgoApplicationPath != "" {
		eh.triggerAppInfraDeploy(stepData.Name, e)
		return nil
	}

	var afterTestsExist bool
	for _, t := range stepData.Tests {
		if !t.ShouldExecuteBefore() {
			afterTestsExist = true
		}
	}
	if afterTestsExist {
		return eh.testHandler.TriggerTest(gitRepoSchemaInfo, stepData, false, gitCommit, e)
	}
	return nil
}

func (eh *eventHandler) handleFailureEvent(e *event.FailureEvent) {
	eh.deploymentLogHandler.MarkStepFailedWithProcessingError(e, e.StepName, e.Error)
}

func (eh *eventHandler) triggerNextSteps(pipelineData *data.PipelineData, step *data.StepData, gitRepoSchemaInfo *git.GitRepoSchemaInfo, e event.Event) error {
	var currentGitCommit string
	pipelineTriggerEvent, isPipelineTriggerEvent := e.(*event.PipelineTriggerEvent)
	if isPipelineTriggerEvent {
		currentGitCommit = pipelineTriggerEvent.RevisionHash
	} else if step.Name != data.RootStepName {
		if err := eh.deploymentLogHandler.MarkStepSuccessful(e, e.GetStepName()); err != nil {
			return err
		}
		currentGitCommitRes, err := eh.deploymentLogHandler.GetCurrentGitCommitHash(e, step.Name)
		if err != nil {
			return err
		}
		currentGitCommit = currentGitCommitRes
	}

	if eh.isSubPipelineRun(e) {
		return nil
	}

	currentPipelineUvn := e.GetUVN()
	childrenSteps := pipelineData.StepChildren[step.Name]
	triggerStepEvents := make([]event.Event, 0)
	for _, stepName := range childrenSteps {
		nextStep := pipelineData.GetStep(stepName)
		parentSteps := pipelineData.StepParents[stepName]
		if eh.deploymentLogHandler.AreParentStepsComplete(e, parentSteps) {
			triggerStepEvents = append(triggerStepEvents, &event.TriggerStepEvent{
				OrgName:       e.GetOrgName(),
				TeamName:      e.GetTeamName(),
				PipelineName:  e.GetPipelineName(),
				StepName:      nextStep.Name,
				UVN:           currentPipelineUvn,
				GitCommitHash: currentGitCommit,
				Rollback:      false,
			})
		}
	}
	// TODO: send message via wq client
	//eh.wqClient.SendMessage(triggerStepEvents)
	return nil
}

func (eh *eventHandler) isSubPipelineRun(e event.Event) bool {
	key := dbkey.MakeDbPipelineInfoKey(e.GetOrgName(), e.GetTeamName(), e.GetPipelineName())
	pipelineInfo := eh.dbClient.FetchLatestPipelineInfo(key)
	// IS a sub-pipeline run, given that only one step is being run and the step names equal each other
	// this could also catch cases where there is a 1-step pipeline, but the behavior will be the same
	return len(pipelineInfo.StepList) == 1 && e.GetStepName() == pipelineInfo.StepList[0]
}

func (eh *eventHandler) triggerAppInfraDeploy(stepName string, e event.Event) {
	// TODO: send message via wq client
	//eh.wqClient.SendMessage(&event.ApplicationInfraTriggerEvent{
	//	OrgName:      e.GetOrgName(),
	//	TeamName:     e.GetTeamName(),
	//	PipelineName: e.GetPipelineName(),
	//	UVN:          e.GetUVN(),
	//	StepName:     stepName,
	//})
}

func (eh *eventHandler) rollback(e event.Event) error {
	hash, err := eh.deploymentLogHandler.GetCurrentGitCommitHash(e, e.GetStepName())
	if err != nil {
		return err
	}
	log.Println("remove this log, ", hash)
	// TODO: send message via wq client
	//eh.wqClient.SendMessage(&event.TriggerStepEvent{
	//	OrgName:       e.GetOrgName(),
	//	TeamName:      e.GetTeamName(),
	//	PipelineName:  e.GetPipelineName(),
	//	StepName:      e.GetStepName(),
	//	UVN:           e.GetUVN(),
	//	GitCommitHash: hash,
	//	Rollback:      true,
	//})
}

func (eh *eventHandler) fetchTeamSchema(e event.Event) *team.TeamSchema {
	key := dbkey.MakeDbTeamKey(e.GetOrgName(), e.GetTeamName())
	schema := eh.dbClient.FetchTeamSchema(key)
	return &schema
}

func (eh *eventHandler) fetchPipelineData(e event.Event, gitRepoSchemaInfo *git.GitRepoSchemaInfo, gitCommitHash string) (*data.PipelineData, error) {
	getFileRequest := &git.GetFileRequest{
		GitRepoSchemaInfo: *gitRepoSchemaInfo,
		Filename:          PipelineFileName,
		GitCommitHash:     gitCommitHash,
	}
	resString, err := eh.repoManagerApi.GetFileFromRepo(getFileRequest, e.GetOrgName(), e.GetTeamName())
	if err != nil {
		return nil, err
	}
	var res data.PipelineData
	if err = yaml.Unmarshal([]byte(resString), &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (eh *eventHandler) getTestNameFromNumber(stepData *data.StepData, testNumber int) string {
	return stepData.Tests[testNumber].GetPath()
}
