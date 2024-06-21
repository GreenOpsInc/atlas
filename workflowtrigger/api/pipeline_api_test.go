package api

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	rdb "github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/pipeline"
	"github.com/greenopsinc/util/team"
	"github.com/rafaeljusto/redigomock/v3"
	"github.com/stretchr/testify/require"
	"greenops.io/workflowtrigger/mocks/db"
	"greenops.io/workflowtrigger/mocks/kubernetesclient"
	"greenops.io/workflowtrigger/serializer"
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
		buildStubs    func(conn *redigomock.Conn)
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
			buildStubs: func(conn *redigomock.Conn) {
				newTeam := team.New(sampleTeamSchema.TeamName, sampleTeamSchema.ParentTeamName, sampleTeamSchema.OrgName)
				listOfTeams := make([]string, 0)

				conn.Command("UNWATCH")
				conn.Command("WATCH", key)
				conn.Command("EXISTS", key).Expect(int64(0))
				conn.Command("MULTI")
				conn.Command("SET", key, serializer.Serialize(newTeam))
				conn.Command("EXEC").Expect(serializer.Serialize(newTeam))

				conn.Command("UNWATCH")
				conn.Command("WATCH", dbkey)
				conn.Command("EXISTS", dbkey).Expect(int64(0))
				conn.Command("MULTI")
				conn.Command("SET", dbkey, serializer.Serialize(append(listOfTeams, sampleTeamSchema.TeamName)))
				conn.Command("EXEC").Expect(serializer.Serialize(append(listOfTeams, sampleTeamSchema.TeamName)))
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
			buildStubs: func(conn *redigomock.Conn) {
				newTeam := team.New(sampleTeamSchema.TeamName, sampleTeamSchema.ParentTeamName, sampleTeamSchema.OrgName)

				conn.Command("UNWATCH")
				conn.Command("WATCH", key)
				conn.Command("EXISTS", key).Expect(int64(1))
				conn.Command("GET", key).Expect(serializer.Serialize(newTeam))
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
			conn := redigomock.NewConn()
			tc.buildStubs(conn)

			operator := db.New(conn)
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
		buildDbStubs  func(conn *redigomock.Conn)
		vars          map[string]string
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			vars: map[string]string{
				"orgName":  "org",
				"teamName": "exampleTeam",
			},
			buildDbStubs: func(conn *redigomock.Conn) {
				newTeam := team.New(sampleTeamSchema.TeamName, sampleTeamSchema.ParentTeamName, sampleTeamSchema.OrgName)
				listOfTeams := make([]string, 0)
				conn.Command("UNWATCH")
				conn.Command("WATCH", key)
				conn.Command("EXISTS", key).Expect(int64(1))
				conn.Command("GET", key).Expect(serializer.Serialize(newTeam))
				conn.Command("MULTI")
				conn.Command("DEL", key)
				conn.Command("EXEC").Expect(serializer.Serialize(nil))

				conn.Command("UNWATCH")
				conn.Command("WATCH", dbkey)
				conn.Command("EXISTS", dbkey).Expect(int64(1))
				conn.Command("GET", dbkey).Expect(serializer.Serialize(append(listOfTeams, sampleTeamSchema.TeamName)))
				conn.Command("MULTI")
				conn.Command("SET", dbkey, serializer.Serialize(listOfTeams))
				conn.Command("EXEC").Expect(serializer.Serialize(listOfTeams))
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
			buildDbStubs: func(conn *redigomock.Conn) {
				conn.Command("UNWATCH")
				conn.Command("WATCH", key)
				conn.Command("EXISTS", key).Expect(int64(1))
				conn.Command("GET", key).Expect(serializer.Serialize(sampleTeamSchema))
				conn.Command("MULTI")
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
			k8sClient := kubernetesclient.New()

			conn := redigomock.NewConn()
			tc.buildDbStubs(conn)
			operator := db.New(conn)

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
	newTeam := team.New(sampleTeamSchema.TeamName, sampleTeamSchema.ParentTeamName, sampleTeamSchema.OrgName)

	testCases := []struct {
		name          string
		buildDbStubs  func(conn *redigomock.Conn)
		vars          map[string]string
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			vars: map[string]string{
				"orgName":  "org",
				"teamName": "exampleTeam",
			},
			buildDbStubs: func(conn *redigomock.Conn) {
				conn.Command("UNWATCH")
				conn.Command("WATCH", key)
				conn.Command("EXISTS", key).Expect(int64(1))
				conn.Command("GET", key).Expect(serializer.Serialize(newTeam))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Result().StatusCode)
				teamSchema := team.UnmarshallTeamSchemaString(getRecorderBody(t, recorder.Body))
				require.Equal(t, teamSchema.TeamName, newTeam.TeamName)
				require.Equal(t, teamSchema.ParentTeamName, newTeam.ParentTeamName)
				require.Equal(t, teamSchema.OrgName, newTeam.OrgName)
			},
		},
		{
			name: "TeamNotFound",
			vars: map[string]string{
				"orgName":  "org",
				"teamName": "exampleTeam",
			},
			buildDbStubs: func(conn *redigomock.Conn) {
				conn.Command("UNWATCH")
				conn.Command("WATCH", key)
				conn.Command("EXISTS", key).Expect(int64(1))
				conn.Command("GET", key).Expect(serializer.Serialize(team.TeamSchema{}))
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
			conn := redigomock.NewConn()
			tc.buildDbStubs(conn)
			operator := db.New(conn)

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
		buildDbStubs  func(conn *redigomock.Conn)
		vars          map[string]string
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "EmptyList",
			vars: map[string]string{
				"orgName": "org",
			},
			buildDbStubs: func(conn *redigomock.Conn) {
				conn.Command("UNWATCH")
				conn.Command("WATCH", key)
				conn.Command("EXISTS", key).Expect(int64(1))
				conn.Command("GET", key).Expect(serializer.Serialize(nil))
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
			buildDbStubs: func(conn *redigomock.Conn) {
				conn.Command("UNWATCH")
				conn.Command("WATCH", key)
				conn.Command("EXISTS", key).Expect(int64(1))
				conn.Command("GET", key).Expect(serializer.Serialize([]string{sampleTeamSchema.TeamName}))
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
			conn := redigomock.NewConn()
			tc.buildDbStubs(conn)
			operator := db.New(conn)

			InitClients(operator, nil, nil, nil, nil, nil, schemaValidator)
			req := httptest.NewRequest(http.MethodPost, "/team/{orgName}/{teamName}", nil)
			req = mux.SetURLVars(req, tc.vars)
			w := httptest.NewRecorder()
			listTeams(w, req)

			tc.checkResponse(t, w)
		})
	}
}

func TestCreatePipeline(t *testing.T) {

}

func getRecorderBody(t *testing.T, body *bytes.Buffer) string {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	return string(data)
}
