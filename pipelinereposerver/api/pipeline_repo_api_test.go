package api

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/greenopsinc/pipelinereposerver/repomanager"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/starter"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	GitRepoUrl       = "https://github.com/GreenOpsInc/atlasexamples"
	PathToRoot       = "basic/"
	ExampleGitCommit = "example-git-commit"
	GitCommit        = "0c586f73b3094bb4f33942809e2f1fa728467d80"
)

var (
	goodGitRepoSchema = git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	badGitRepoSchema = git.GitRepoSchema{
		GitRepo:    "testing",
		PathToRoot: "testing",
		GitCred:    nil,
	}
	gitRepoSchemaInfo = git.GitRepoSchemaInfo{
		GitRepoUrl: GitRepoUrl,
		PathToRoot: PathToRoot,
	}
)

func init() {
	var dbOperator db.DbOperator
	dbOperator = db.New(starter.GetDbClientConfig())
	repoManager = repomanager.New(dbOperator, nil, "org")
}

func TestCloneRepoNoMuxVariables(t *testing.T) {
	body, _ := json.Marshal(goodGitRepoSchema)
	req := httptest.NewRequest(http.MethodPost, "/repo/clone/{orgName}", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	cloneRepo(w, req)
	if want, got := http.StatusBadRequest, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestCloneRepoBadBodyRequestShouldPanic(t *testing.T) {
	vars := map[string]string{
		"orgName": "org",
	}
	req := httptest.NewRequest(http.MethodPost, "/repo/clone/{orgName}", strings.NewReader("////"))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("request did not panic")
		}
	}()
	cloneRepo(w, req)
}

func TestCloneRepoOrgNameNotFound(t *testing.T) {
	vars := map[string]string{
		"orgName": "some_random_org",
	}
	body, _ := json.Marshal(goodGitRepoSchema)
	req := httptest.NewRequest(http.MethodPost, "/repo/clone/{orgName}", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	cloneRepo(w, req)
	if want, got := http.StatusBadRequest, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestCloneRepoBadGitRepoSchema(t *testing.T) {
	vars := map[string]string{
		"orgName": "org",
	}
	body, _ := json.Marshal(badGitRepoSchema)
	req := httptest.NewRequest(http.MethodPost, "/repo/clone/{orgName}", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("request did not panic")
		}
	}()
	cloneRepo(w, req)
}

func TestCloneRepoGoodGitRepoSchemaNotFound(t *testing.T) {
	vars := map[string]string{
		"orgName": "org",
	}
	body, _ := json.Marshal(goodGitRepoSchema)
	req := httptest.NewRequest(http.MethodPost, "/repo/clone/{orgName}", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	cloneRepo(w, req)
	if want, got := http.StatusOK, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestCloneRepoGitRepoSchemaExist(t *testing.T) {
	vars := map[string]string{
		"orgName": "org",
	}
	body, _ := json.Marshal(goodGitRepoSchema)
	req := httptest.NewRequest(http.MethodPost, "/repo/clone/{orgName}", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()

	if isCloned := repoManager.Clone(goodGitRepoSchema); !isCloned {
		t.Fatalf("Cloning repo %s failed.", goodGitRepoSchema.GitRepo)
	}

	cloneRepo(w, req)
	if want, got := http.StatusOK, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestDeleteRepoBadGitRepoSchema(t *testing.T) {
	body, _ := json.Marshal(badGitRepoSchema)
	req := httptest.NewRequest(http.MethodPost, "/repo/delete", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("request did not panic")
		}
	}()
	deleteRepo(w, req)
}

func TestDeleteRepoGitRepoSchemaNotFound(t *testing.T) {
	body, _ := json.Marshal(goodGitRepoSchema)
	req := httptest.NewRequest(http.MethodPost, "/repo/delete", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	deleteRepo(w, req)
	if want, got := http.StatusOK, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestDeleteRepoGitRepoSchemaExist(t *testing.T) {
	body, _ := json.Marshal(goodGitRepoSchema)
	req := httptest.NewRequest(http.MethodPost, "/repo/delete", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	if isCloned := repoManager.Clone(goodGitRepoSchema); !isCloned {
		t.Fatalf("Cloning repo %s failed.", goodGitRepoSchema.GitRepo)
	}

	deleteRepo(w, req)
	if want, got := http.StatusOK, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestSyncRepoBadGitRepoSchema(t *testing.T) {
	body, _ := json.Marshal(badGitRepoSchema)
	req := httptest.NewRequest(http.MethodPost, "/repo/sync", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("request did not panic")
		}
	}()
	syncRepo(w, req)
}

func TestSyncRepoGitRepoSchemaNotFound(t *testing.T) {
	body, _ := json.Marshal(goodGitRepoSchema)
	req := httptest.NewRequest(http.MethodPost, "/repo/sync", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	syncRepo(w, req)
	if want, got := http.StatusInternalServerError, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestSyncRepoGitRepoSchemaExist(t *testing.T) {
	body, _ := json.Marshal(goodGitRepoSchema)
	req := httptest.NewRequest(http.MethodPost, "/repo/sync", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	if isCloned := repoManager.Clone(goodGitRepoSchema); !isCloned {
		t.Fatalf("Cloning repo %s failed.", goodGitRepoSchema.GitRepo)
	}
	syncRepo(w, req)
	if want, got := http.StatusOK, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestResetRepoToVersionNoMuxVars(t *testing.T) {
	body, _ := json.Marshal(gitRepoSchemaInfo)
	req := httptest.NewRequest(http.MethodPost, "/repo/resetToVersion/{orgName}/{teamName}/{gitCommit}", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	resetRepoToVersion(w, req)
	if want, got := http.StatusBadRequest, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestResetRepoToVersionBadGitRepoSchema(t *testing.T) {
	vars := map[string]string{
		"orgName":   "org",
		"teamName":  "exampleTeam",
		"gitCommit": ExampleGitCommit,
	}
	body, _ := json.Marshal(badGitRepoSchema)
	req := httptest.NewRequest(http.MethodPost, "/repo/resetToVersion/{orgName}/{teamName}/{gitCommit}", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("request did not panic")
		}
	}()
	resetRepoToVersion(w, req)
}

func TestResetRepoToVersionGitRepoSchemaNotFound(t *testing.T) {
	vars := map[string]string{
		"orgName":   "org",
		"teamName":  "exampleTeam",
		"gitCommit": ExampleGitCommit,
	}

	body, _ := json.Marshal(gitRepoSchemaInfo)
	req := httptest.NewRequest(http.MethodPost, "/repo/resetToVersion/{orgName}/{teamName}/{gitCommit}", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	resetRepoToVersion(w, req)
	if want, got := http.StatusBadRequest, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestResetRepoToVersionOrgNameNotMatch(t *testing.T) {
	vars := map[string]string{
		"orgName":   "test",
		"teamName":  "exampleTeam",
		"gitCommit": ExampleGitCommit,
	}
	body, _ := json.Marshal(gitRepoSchemaInfo)
	req := httptest.NewRequest(http.MethodPost, "/repo/resetToVersion/{orgName}/{teamName}/{gitCommit}", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	resetRepoToVersion(w, req)
	if want, got := http.StatusBadRequest, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestResetRepoToVersionFailed(t *testing.T) {
	vars := map[string]string{
		"orgName":   "org",
		"teamName":  "exampleTeam",
		"gitCommit": ExampleGitCommit,
	}
	if isCloned := repoManager.Clone(goodGitRepoSchema); !isCloned {
		t.Fatalf("Cloning repo %s failed.", goodGitRepoSchema.GitRepo)
	}

	body, _ := json.Marshal(gitRepoSchemaInfo)
	req := httptest.NewRequest(http.MethodPost, "/repo/resetToVersion/{orgName}/{teamName}/{gitCommit}", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	resetRepoToVersion(w, req)
	if want, got := http.StatusInternalServerError, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestResetRepoToVersionSuccess(t *testing.T) {
	vars := map[string]string{
		"orgName":   "org",
		"teamName":  "exampleTeam",
		"gitCommit": GitCommit,
	}
	if isCloned := repoManager.Clone(goodGitRepoSchema); !isCloned {
		t.Fatalf("Cloning repo %s failed.", goodGitRepoSchema.GitRepo)
	}

	body, _ := json.Marshal(gitRepoSchemaInfo)
	req := httptest.NewRequest(http.MethodPost, "/repo/resetToVersion/{orgName}/{teamName}/{gitCommit}", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	resetRepoToVersion(w, req)
	if want, got := http.StatusOK, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}
