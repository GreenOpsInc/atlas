package reposerver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/httpclient"
	"github.com/greenopsinc/util/tlsmanager"
	"greenops.io/workflowtrigger/serializer"
)

const (
	RootCommit       string = "ROOT_COMMIT"
	PipelineFileName string = "pipeline.yaml"
	getFileExtension string = "file"
)

type RepoManagerApi interface {
	CloneRepo(orgName string, gitRepoSchema git.GitRepoSchema) bool
	DeleteRepo(gitRepoSchema git.GitRepoSchema) bool
	UpdateRepo(orgName string, oldGitRepoSchema git.GitRepoSchema, newGitRepoSchema git.GitRepoSchema) bool
	SyncRepo(gitRepoSchema git.GitRepoSchema) string
	GetFileFromRepo(getFileRequest git.GetFileRequest, orgName string, teamName string) string
}

type RepoManagerApiImpl struct {
	serverEndpoint string
	client         httpclient.HttpClient
}

func New(serverEndpoint string, tm tlsmanager.Manager) (RepoManagerApi, error) {
	if strings.HasSuffix(serverEndpoint, "/") {
		serverEndpoint = serverEndpoint[:len(serverEndpoint)-1]
	}
	httpClient, err := httpclient.New(tlsmanager.ClientRepoServer, tm)
	if err != nil {
		return nil, err
	}
	return &RepoManagerApiImpl{
		serverEndpoint: serverEndpoint,
		client:         httpClient,
	}, nil
}

func (r *RepoManagerApiImpl) CloneRepo(orgName string, gitRepoSchema git.GitRepoSchema) bool {
	var err error
	var payload []byte
	var request *http.Request
	payload = []byte(serializer.Serialize(gitRepoSchema))
	request, err = http.NewRequest("POST", r.serverEndpoint+"/repo/clone/"+orgName, bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	log.Printf("Clone repo request returned status code %d - %s", resp.StatusCode, buf.String())
	return resp.StatusCode == 200
}

func (r *RepoManagerApiImpl) DeleteRepo(gitRepoSchema git.GitRepoSchema) bool {
	var err error
	var payload []byte
	var request *http.Request
	payload = []byte(serializer.Serialize(gitRepoSchema))
	request, err = http.NewRequest("POST", r.serverEndpoint+"/repo/delete", bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	log.Printf("Delete repo request returned status code %d - %s", resp.StatusCode, buf.String())
	return resp.StatusCode == 200
}

func (r *RepoManagerApiImpl) UpdateRepo(orgName string, oldGitRepoSchema git.GitRepoSchema, newGitRepoSchema git.GitRepoSchema) bool {
	if deleted := r.DeleteRepo(oldGitRepoSchema); !deleted {
		return false
	}
	return r.CloneRepo(orgName, newGitRepoSchema)
}

func (r *RepoManagerApiImpl) SyncRepo(gitRepoSchema git.GitRepoSchema) string {
	var err error
	var payload []byte
	var request *http.Request
	payload = []byte(serializer.Serialize(gitRepoSchema))
	request, err = http.NewRequest("POST", r.serverEndpoint+"/repo/sync", bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	log.Printf("Sync repo request returned status code %d", resp.StatusCode)
	if resp.StatusCode == 200 {
		revisionHash, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(fmt.Sprintf("unexpected err %s", err.Error()))
		}
		return string(revisionHash)
	} else {
		return ""
	}
}

func (r *RepoManagerApiImpl) GetFileFromRepo(getFileRequest git.GetFileRequest, orgName string, teamName string) string {
	var err error
	var payload []byte
	var request *http.Request
	payload, _ = json.Marshal(getFileRequest)
	request, err = http.NewRequest("POST", r.serverEndpoint+fmt.Sprintf("/data/%s/%s/%s", getFileExtension, orgName, teamName), bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	log.Printf("Get file from repo request returned status code %d", resp.StatusCode)
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic(buf.String())
	}
	return buf.String()
}
