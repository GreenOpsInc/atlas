package api

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/greenopsinc/pipelinereposerver/repomanager"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/kubernetesclient"
	"net/http"
	"os"
)

const (
	orgNameField string = "orgName"
	teamNameField string = "teamName"
)

var repoManager repomanager.RepoManager

func cloneRepo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	var gitRepoSchema git.GitRepoSchema
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	gitRepoSchema = git.UnmarshallGitRepoSchemaString(string(buf.Bytes()))
	if repoManager.GetOrgName() != orgName {
		http.Error(w, "Org names don't match", http.StatusBadRequest)
		return
	}
	if repoManager.ContainsGitRepoSchema(gitRepoSchema) {
		if repoManager.Update(gitRepoSchema) {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			http.Error(w, "Git repo schema could not be read", http.StatusBadRequest)
			return
		}
	} else {
		if repoManager.Clone(gitRepoSchema) {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			http.Error(w, "Internal server error occurred", http.StatusInternalServerError)
			return
		}
	}
}

func deleteRepo(w http.ResponseWriter, r *http.Request) {
	var gitRepoSchema git.GitRepoSchema
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	gitRepoSchema = git.UnmarshallGitRepoSchemaString(string(buf.Bytes()))
	if repoManager.Delete(gitRepoSchema) {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		http.Error(w, "Internal server error occurred", http.StatusInternalServerError)
		return
	}
}

func syncRepo(w http.ResponseWriter, r *http.Request) {
	var gitRepoSchema git.GitRepoSchema
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	gitRepoSchema = git.UnmarshallGitRepoSchemaString(string(buf.Bytes()))
	revisionHash := repoManager.Sync(gitRepoSchema)
	if revisionHash != "" {
		w.Write([]byte(revisionHash))
		return
	} else {
		http.Error(w, "Internal server error occurred", http.StatusInternalServerError)
		return
	}
}

func resetRepoToVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	//teamName := vars[teamNameField]
	gitCommit := vars["gitCommit"]
	var gitRepoSchemaInfo git.GitRepoSchemaInfo
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	gitRepoSchemaInfo = git.UnmarshallGitRepoSchemaInfoString(string(buf.Bytes()))
	gitRepoSchema := git.New(gitRepoSchemaInfo.GetGitRepo(), gitRepoSchemaInfo.GetPathToRoot(), nil)
	if repoManager.GetOrgName() == orgName && repoManager.ContainsGitRepoSchema(gitRepoSchema) {
		if repoManager.ResetToVersion(gitCommit, gitRepoSchema) {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			http.Error(w, "Internal server error occurred", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Org name doesn't match or repo is not contained by manager", http.StatusBadRequest)
		return
	}
}

func InitClients(dbOperator db.DbOperator, kubernetesClient kubernetesclient.KubernetesClient) {
	orgName := os.Getenv("ORG_NAME")
	if orgName == "" {
		orgName = repomanager.DefaultOrg
	}
	repoManager = repomanager.New(dbOperator, kubernetesClient, orgName)
}

func InitRepoEndpoints(r *mux.Router) {
	r.HandleFunc("/repo/clone/{orgName}", cloneRepo).Methods("POST")
	r.HandleFunc("/repo/delete", deleteRepo).Methods("POST")
	r.HandleFunc("/repo/sync", syncRepo).Methods("POST")
	r.HandleFunc("/repo/resetToVersion/{orgName}/{teamName}/{gitCommit}", resetRepoToVersion).Methods("POST")
}
