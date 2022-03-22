package handling

import (
	"log"

	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/event"
	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/pipeline/data"
	cw "github.com/greenopsinc/workfloworchestrator/ingest/apiclient/clientwrapper"
	rs "github.com/greenopsinc/workfloworchestrator/ingest/apiclient/reposerver"
)

type TestHandler interface {
	TriggerTest(gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, beforeTest bool, gitCommitHash string, e event.Event) error
	CreateAndRunTest(clusterName string, stepData *data.StepData, gitRepoSchemaInfo *git.GitRepoSchemaInfo, test data.TestData, testNumber int, gitCommitHash string, e event.Event) error
}

type testHandler struct {
	repoManagerAPI     rs.RepoManagerAPI
	clientRequestQueue cw.ClientRequestQueue
}

func NewTestHandler(repoManagerAPI rs.RepoManagerAPI, clientRequestQueue cw.ClientRequestQueue) TestHandler {
	return &testHandler{
		repoManagerAPI:     repoManagerAPI,
		clientRequestQueue: clientRequestQueue,
	}
}

func (t *testHandler) TriggerTest(gitRepoSchemaInfo *git.GitRepoSchemaInfo, stepData *data.StepData, beforeTest bool, gitCommitHash string, e event.Event) error {
	for i := 0; i < len(stepData.Tests); i++ {
		if beforeTest == stepData.Tests[i].ShouldExecuteBefore() {
			return t.CreateAndRunTest(stepData.ClusterName, stepData, gitRepoSchemaInfo, stepData.Tests[i], i, gitCommitHash, e)
		}
	}
	return nil
}

func (t *testHandler) CreateAndRunTest(clusterName string, stepData *data.StepData, gitRepoSchemaInfo *git.GitRepoSchemaInfo, test data.TestData, testNumber int, gitCommitHash string, e event.Event) error {
	getFileRequest := &git.GetFileRequest{
		GitRepoSchemaInfo: *gitRepoSchemaInfo,
		Filename:          test.GetPath(),
		GitCommitHash:     gitCommitHash,
	}
	testConfig, err := t.repoManagerAPI.GetFileFromRepo(getFileRequest, e.GetOrgName(), e.GetTeamName())
	if err != nil {
		return err
	}

	log.Println("creating test Job...")
	stepNS, err := GetStepNamespace(e, t.repoManagerAPI, stepData.ArgoApplicationPath, gitRepoSchemaInfo, gitCommitHash)
	if err != nil {
		return err
	}
	return t.clientRequestQueue.DeployAndWatch(
		clusterName,
		e.GetOrgName(),
		e.GetTeamName(),
		e.GetPipelineName(),
		e.GetUvn(),
		stepData.Name,
		stepNS,
		cw.DeployTestRequest,
		clientrequest.LatestRevision,
		test.GetPayload(testNumber, testConfig),
		test.GetWatchKey(),
		testNumber,
	)
}
