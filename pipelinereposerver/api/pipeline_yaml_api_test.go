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
	Filename  = "pipeline.yaml"
	Filename2 = "testing"
)

var (
	goodGetFileRequest = &git.GetFileRequest{
		GitRepoSchemaInfo: gitRepoSchemaInfo,
		Filename:          Filename,
		GitCommitHash:     GitCommit,
	}
	badGetFileRequest = &git.GetFileRequest{
		GitRepoSchemaInfo: gitRepoSchemaInfo,
		Filename:          Filename,
		GitCommitHash:     ExampleGitCommit,
	}
	badGetFileRequest2 = &git.GetFileRequest{
		GitRepoSchemaInfo: gitRepoSchemaInfo,
		Filename:          Filename2,
		GitCommitHash:     GitCommit,
	}
)

func init() {
	var dbOperator db.DbOperator
	dbOperator = db.New(starter.GetDbClientConfig())
	repoManager = repomanager.New(dbOperator, nil, "org")
}

func TestGetPipelineConfigBadBody(t *testing.T) {
	vars := map[string]string{
		"orgName":  "org",
		"teamName": "exampleTeam",
	}
	req := httptest.NewRequest(http.MethodPost, "/data/file/{orgName}/{teamName}", strings.NewReader("////"))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	getPipelineConfig(w, req)
	if want, got := http.StatusBadRequest, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestGetPipelineConfigNoMuxVars(t *testing.T) {
	body, _ := json.Marshal(goodGetFileRequest)
	req := httptest.NewRequest(http.MethodPost, "/data/file/{orgName}/{teamName}", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	getPipelineConfig(w, req)
	if want, got := http.StatusBadRequest, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestGetPipelineConfigOrgNameNotMatched(t *testing.T) {
	vars := map[string]string{
		"orgName":  "test",
		"teamName": "exampleTeam",
	}
	body, _ := json.Marshal(goodGetFileRequest)
	req := httptest.NewRequest(http.MethodPost, "/data/file/{orgName}/{teamName}", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	getPipelineConfig(w, req)
	if want, got := http.StatusBadRequest, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestGetPipelineConfigGitRepoSchemaNotFound(t *testing.T) {
	vars := map[string]string{
		"orgName":  "org",
		"teamName": "exampleTeam",
	}
	body, _ := json.Marshal(goodGetFileRequest)
	req := httptest.NewRequest(http.MethodPost, "/data/file/{orgName}/{teamName}", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	getPipelineConfig(w, req)
	if want, got := http.StatusBadRequest, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestGetPipelineConfigResetToVersionFailed(t *testing.T) {
	vars := map[string]string{
		"orgName":  "org",
		"teamName": "exampleTeam",
	}
	if isCloned := repoManager.Clone(goodGitRepoSchema); !isCloned {
		t.Fatalf("Cloning repo %s failed.", goodGitRepoSchema.GitRepo)
	}

	body, _ := json.Marshal(badGetFileRequest)
	req := httptest.NewRequest(http.MethodPost, "/data/file/{orgName}/{teamName}", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	getPipelineConfig(w, req)
	if want, got := http.StatusInternalServerError, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestGetPipelineConfigGetYamFileContentsFailed(t *testing.T) {
	vars := map[string]string{
		"orgName":  "org",
		"teamName": "exampleTeam",
	}
	if isCloned := repoManager.Clone(goodGitRepoSchema); !isCloned {
		t.Fatalf("Cloning repo %s failed.", goodGitRepoSchema.GitRepo)
	}

	body, _ := json.Marshal(badGetFileRequest2)
	req := httptest.NewRequest(http.MethodPost, "/data/file/{orgName}/{teamName}", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	getPipelineConfig(w, req)
	if want, got := http.StatusNotFound, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestGetPipelineConfigSuccess(t *testing.T) {
	vars := map[string]string{
		"orgName":  "org",
		"teamName": "exampleTeam",
	}
	if isCloned := repoManager.Clone(goodGitRepoSchema); !isCloned {
		t.Fatalf("Cloning repo %s failed.", goodGitRepoSchema.GitRepo)
	}

	body, _ := json.Marshal(goodGetFileRequest)
	req := httptest.NewRequest(http.MethodPost, "/data/file/{orgName}/{teamName}", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	getPipelineConfig(w, req)
	if want, got := http.StatusOK, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}
