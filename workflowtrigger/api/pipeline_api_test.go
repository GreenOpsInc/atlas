package api

import (
	"bytes"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	rdb "github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/pipeline"
	"github.com/greenopsinc/util/team"
	"github.com/stretchr/testify/require"
	"greenops.io/workflowtrigger/mocks/db"
	"greenops.io/workflowtrigger/mocks/kubernetesclient"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	GitRepoUrl = "https://github.com/GreenOpsInc/atlasexamples"
	PathToRoot = "basic/"
)

var (
	sampleGitRepoSchema = git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	samplePipeline = pipeline.PipelineSchema{
		PipelineName:  "examplePipeline",
		GitRepoSchema: sampleGitRepoSchema,
	}
	sampleTeamSchema = team.TeamSchema{
		TeamName:       "exampleTeam",
		ParentTeamName: "na",
		OrgName:        "org",
		Pipelines:      []*pipeline.PipelineSchema{&samplePipeline},
	}
)

func TestCreateTeam(t *testing.T) {
	key := rdb.MakeDbTeamKey(sampleTeamSchema.OrgName, sampleTeamSchema.TeamName)
	dbkey := rdb.MakeDbListOfTeamsKey(sampleTeamSchema.OrgName)

	testCases := []struct {
		name          string
		buildStubs    func(operator *db.MockDbOperator, ctrl *gomock.Controller)
		vars          map[string]string
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			vars: map[string]string{
				"orgName":        "org",
				"teamName":       "exampleTeam",
				"parentTeamName": "na",
			},
			buildStubs: func(operator *db.MockDbOperator, ctrl *gomock.Controller) {
				newTeam := team.New(sampleTeamSchema.TeamName, sampleTeamSchema.ParentTeamName, sampleTeamSchema.OrgName)
				listOfTeams := make([]string, 0)

				dbClient := db.NewMockDbClient(ctrl)
				operator.EXPECT().GetClient().Times(1).Return(dbClient)
				dbClient.EXPECT().FetchTeamSchema(key).Times(1).Return(team.TeamSchema{})
				dbClient.EXPECT().FetchStringList(dbkey).Times(1).Return(listOfTeams)
				dbClient.EXPECT().Close().Times(1).Return()
				gomock.InOrder(
					dbClient.EXPECT().StoreValue(key, newTeam).Times(1).Return(),
					dbClient.EXPECT().StoreValue(dbkey, append(listOfTeams, sampleTeamSchema.TeamName)).Times(1).Return(),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Result().StatusCode)
			},
		},
		{
			name: "TeamAlreadyExists",
			vars: map[string]string{
				"orgName":        "org",
				"teamName":       "exampleTeam",
				"parentTeamName": "na",
			},
			buildStubs: func(operator *db.MockDbOperator, ctrl *gomock.Controller) {
				dbClient := db.NewMockDbClient(ctrl)
				operator.EXPECT().GetClient().Times(1).Return(dbClient)
				dbClient.EXPECT().FetchTeamSchema(key).Times(1).Return(sampleTeamSchema)
				dbClient.EXPECT().Close().Times(1).Return()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode)
				require.Equal(t, "team already exists\n", getRecorderBody(t, recorder.Body))
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			operator := db.NewMockDbOperator(ctrl)
			tc.buildStubs(operator, ctrl)

			InitClients(operator, nil, nil, nil, nil, nil, schemaValidator)
			req := httptest.NewRequest(http.MethodPost, "/team/{orgName}/{parentTeamName}/{teamName}", nil)
			req = mux.SetURLVars(req, tc.vars)
			w := httptest.NewRecorder()
			createTeam(w, req)

			tc.checkResponse(t, w)
		})
	}
}

func TestDeleteTeam(t *testing.T) {
	key := rdb.MakeDbTeamKey(sampleTeamSchema.OrgName, sampleTeamSchema.TeamName)
	dbkey := rdb.MakeDbListOfTeamsKey(sampleTeamSchema.OrgName)

	testCases := []struct {
		name          string
		buildDbStubs  func(operator *db.MockDbOperator, ctrl *gomock.Controller)
		buildK8sStubs func(client *kubernetesclient.MockKubernetesClient)
		vars          map[string]string
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			vars: map[string]string{
				"orgName":  "org",
				"teamName": "exampleTeam",
			},
			buildDbStubs: func(operator *db.MockDbOperator, ctrl *gomock.Controller) {
				listOfTeams := make([]string, 0)
				dbClient := db.NewMockDbClient(ctrl)
				operator.EXPECT().GetClient().Times(1).Return(dbClient)
				dbClient.EXPECT().FetchTeamSchema(key).Times(1).Return(sampleTeamSchema)
				dbClient.EXPECT().FetchStringList(dbkey).Times(1).Return(listOfTeams)
				dbClient.EXPECT().Close().Times(1).Return()
				gomock.InOrder(
					dbClient.EXPECT().StoreValue(key, nil).Times(1).Return(),
					dbClient.EXPECT().StoreValue(dbkey, listOfTeams).Times(1).Return(),
				)
			},
			buildK8sStubs: func(client *kubernetesclient.MockKubernetesClient) {
				client.EXPECT().
					StoreGitCred(nil, rdb.MakeSecretName(sampleTeamSchema.OrgName, sampleTeamSchema.TeamName, samplePipeline.GetPipelineName())).
					Times(1).
					Return(false)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Result().StatusCode)
			},
		},
		{
			name: "SecretDeletionFailed",
			vars: map[string]string{
				"orgName":  "org",
				"teamName": "exampleTeam",
			},
			buildDbStubs: func(operator *db.MockDbOperator, ctrl *gomock.Controller) {
				dbClient := db.NewMockDbClient(ctrl)
				operator.EXPECT().GetClient().Times(1).Return(dbClient)
				dbClient.EXPECT().FetchTeamSchema(key).Times(1).Return(sampleTeamSchema)
				dbClient.EXPECT().Close().Times(1).Return()
			},
			buildK8sStubs: func(client *kubernetesclient.MockKubernetesClient) {
				client.EXPECT().
					StoreGitCred(nil, rdb.MakeSecretName(sampleTeamSchema.OrgName, sampleTeamSchema.TeamName, samplePipeline.GetPipelineName())).
					Times(1).
					Return(true)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Result().StatusCode)
				require.Equal(t, "kubernetes secret deletion did not work\n", getRecorderBody(t, recorder.Body))
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			k8sClient := kubernetesclient.NewMockKubernetesClient(ctrl)
			operator := db.NewMockDbOperator(ctrl)
			tc.buildDbStubs(operator, ctrl)
			tc.buildK8sStubs(k8sClient)

			InitClients(operator, nil, k8sClient, nil, nil, nil, schemaValidator)
			req := httptest.NewRequest(http.MethodPost, "/team/{orgName}/{teamName}", nil)
			req = mux.SetURLVars(req, tc.vars)
			w := httptest.NewRecorder()
			deleteTeam(w, req)

			tc.checkResponse(t, w)
		})
	}
}

func TestReadTeam(t *testing.T) {
	key := rdb.MakeDbTeamKey(sampleTeamSchema.OrgName, sampleTeamSchema.TeamName)

	testCases := []struct {
		name          string
		buildDbStubs  func(operator *db.MockDbOperator, ctrl *gomock.Controller)
		vars          map[string]string
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			vars: map[string]string{
				"orgName":  "org",
				"teamName": "exampleTeam",
			},
			buildDbStubs: func(operator *db.MockDbOperator, ctrl *gomock.Controller) {
				dbClient := db.NewMockDbClient(ctrl)
				operator.EXPECT().GetClient().Times(1).Return(dbClient)
				dbClient.EXPECT().FetchTeamSchema(key).Times(1).Return(sampleTeamSchema)
				dbClient.EXPECT().Close().Times(1).Return()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Result().StatusCode)
				teamSchema := team.UnmarshallTeamSchemaString(getRecorderBody(t, recorder.Body))
				require.Equal(t, teamSchema.TeamName, sampleTeamSchema.TeamName)
				require.Equal(t, teamSchema.ParentTeamName, sampleTeamSchema.ParentTeamName)
				require.Equal(t, teamSchema.OrgName, sampleTeamSchema.OrgName)
			},
		},
		{
			name: "TeamNotFound",
			vars: map[string]string{
				"orgName":  "org",
				"teamName": "exampleTeam",
			},
			buildDbStubs: func(operator *db.MockDbOperator, ctrl *gomock.Controller) {
				dbClient := db.NewMockDbClient(ctrl)
				operator.EXPECT().GetClient().Times(1).Return(dbClient)
				dbClient.EXPECT().FetchTeamSchema(key).Times(1).Return(team.TeamSchema{})
				dbClient.EXPECT().Close().Times(1).Return()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode)
				require.Equal(t, "no team found\n", getRecorderBody(t, recorder.Body))
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			operator := db.NewMockDbOperator(ctrl)
			tc.buildDbStubs(operator, ctrl)

			InitClients(operator, nil, nil, nil, nil, nil, schemaValidator)
			req := httptest.NewRequest(http.MethodPost, "/team/{orgName}/{teamName}", nil)
			req = mux.SetURLVars(req, tc.vars)
			w := httptest.NewRecorder()
			readTeam(w, req)

			tc.checkResponse(t, w)
		})
	}
}

func TestListTeams(t *testing.T) {
	key := rdb.MakeDbListOfTeamsKey(sampleTeamSchema.OrgName)

	testCases := []struct {
		name          string
		buildDbStubs  func(operator *db.MockDbOperator, ctrl *gomock.Controller)
		vars          map[string]string
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "EmptyList",
			vars: map[string]string{
				"orgName": "org",
			},
			buildDbStubs: func(operator *db.MockDbOperator, ctrl *gomock.Controller) {
				dbClient := db.NewMockDbClient(ctrl)
				operator.EXPECT().GetClient().Times(1).Return(dbClient)
				dbClient.EXPECT().FetchStringList(key).Times(1).Return(make([]string, 0))
				dbClient.EXPECT().Close().Times(1).Return()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Result().StatusCode)
				require.Equal(t, "[]", getRecorderBody(t, recorder.Body))
			},
		},
		{
			name: "WithTeams",
			vars: map[string]string{
				"orgName": "org",
			},
			buildDbStubs: func(operator *db.MockDbOperator, ctrl *gomock.Controller) {
				dbClient := db.NewMockDbClient(ctrl)
				operator.EXPECT().GetClient().Times(1).Return(dbClient)
				dbClient.EXPECT().FetchStringList(key).Times(1).Return([]string{sampleTeamSchema.TeamName})
				dbClient.EXPECT().Close().Times(1).Return()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Result().StatusCode)
				require.Equal(t, fmt.Sprintf("[\"%s\"]", sampleTeamSchema.TeamName), getRecorderBody(t, recorder.Body))
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			operator := db.NewMockDbOperator(ctrl)
			tc.buildDbStubs(operator, ctrl)

			InitClients(operator, nil, nil, nil, nil, nil, schemaValidator)
			req := httptest.NewRequest(http.MethodPost, "/team/{orgName}/{teamName}", nil)
			req = mux.SetURLVars(req, tc.vars)
			w := httptest.NewRecorder()
			listTeams(w, req)

			tc.checkResponse(t, w)
		})
	}
}

func getRecorderBody(t *testing.T, body *bytes.Buffer) string {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	return string(data)
}
