package api

import (
	"greenops.io/workflowtrigger/mocks/db"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTeamSuccess(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/team/org/na/exampleTeam", nil)
	w := httptest.NewRecorder()

	dbOperator = &db.MockRedisClientOperator{}
	createTeam(w, req)

}
