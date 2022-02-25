package api

import (
	"github.com/gorilla/mux"
	"greenops.io/workflowtrigger/mocks/db"
	"greenops.io/workflowtrigger/mocks/kubernetesclient"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	dbOperator = db.New()
	kubernetesClient = kubernetesclient.New()
}

func TestCreateTeam(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/team/{orgName}/{parentTeamName}/{teamName}", nil)
	w := httptest.NewRecorder()

	// Happy team creation
	vars := map[string]string{
		"orgName":        "org",
		"teamName":       "exampleTeam",
		"parentTeamName": "na",
	}
	req = mux.SetURLVars(req, vars)
	createTeam(w, req)
	if want, got := http.StatusOK, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}

	// This must fail, since we already have exampleTeam team inside the in-state DB
	req2 := httptest.NewRequest(http.MethodPost, "/team/{orgName}/{parentTeamName}/{teamName}", nil)
	req2 = mux.SetURLVars(req2, vars)
	w2 := httptest.NewRecorder()
	createTeam(w2, req2)
	if want, got := http.StatusBadRequest, w2.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}

func TestCreatePipeline(t *testing.T) {

}

func TestDeleteTeam(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/team/{orgName}/{teamName}", nil)
	w := httptest.NewRecorder()

	// Happy team deletion
	vars := map[string]string{
		"orgName":  "org",
		"teamName": "exampleTeam",
	}
	req = mux.SetURLVars(req, vars)
	deleteTeam(w, req)
	if want, got := http.StatusOK, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}
}
