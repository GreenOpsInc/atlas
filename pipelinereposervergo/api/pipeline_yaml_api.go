package api

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/greenopsinc/util/git"
	"net/http"
)

const rootCommit string = "ROOT_COMMIT"

func getPipelineConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars[orgNameField]
	//teamName := vars[teamNameField]
	var fileRequest git.GetFileRequest
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal([]byte(buf.String()), &fileRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	gitRepoSchema := git.New(fileRequest.GitRepoSchemaInfo.GetGitRepo(), fileRequest.GitRepoSchemaInfo.GetPathToRoot(), nil)
	if repoManager.GetOrgName() == orgName && repoManager.ContainsGitRepoSchema(gitRepoSchema) {
		//gitCommitHash should never be ROOT_COMMIT
		desiredGitCommit := fileRequest.GitCommitHash
		if desiredGitCommit == rootCommit {
			desiredGitCommit = repoManager.GetCurrentCommit(gitRepoSchema.GetGitRepo())
		}
		if !repoManager.ResetToVersion(desiredGitCommit, gitRepoSchema) {
			http.Error(w, "Could not switch to right revision", http.StatusInternalServerError)
			return
		}
		fileContents := repoManager.GetYamlFileContents(fileRequest.Filename, gitRepoSchema)
		if fileContents == "" {
			http.Error(w, "Couldn't find contents", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "org name did not match or repo is not contained by manager", http.StatusBadRequest)
	return
}

func InitFileEndpoints(r *mux.Router) {
	r.HandleFunc("/data/file/{orgName}/{teamName}", getPipelineConfig).Methods("POST")
}