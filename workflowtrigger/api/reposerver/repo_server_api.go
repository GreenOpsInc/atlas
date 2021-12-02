package reposerver

import (
	"bytes"
	"greenops.io/workflowtrigger/util/git"
	"greenops.io/workflowtrigger/util/serializer"
	"log"
	"net/http"
	"strings"
	"time"
)

type RepoManagerApi interface {
	CloneRepo(orgName string, gitRepoSchema git.GitRepoSchema) bool
	DeleteRepo(gitRepoSchema git.GitRepoSchema) bool
	SyncRepo(gitRepoSchema git.GitRepoSchema) bool
}

type RepoManagerApiImpl struct {
	serverEndpoint string
	client         *http.Client
}

func New(serverEndpoint string) RepoManagerApi {
	if strings.HasSuffix(serverEndpoint, "/") {
		serverEndpoint = serverEndpoint + "repo"
	} else {
		serverEndpoint = serverEndpoint + "/repo"
	}
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	return &RepoManagerApiImpl{
		serverEndpoint: serverEndpoint,
		client:         httpClient,
	}
}

func (r *RepoManagerApiImpl) CloneRepo(orgName string, gitRepoSchema git.GitRepoSchema) bool {
	var err error
	var payload []byte
	var request *http.Request
	payload = []byte(serializer.Serialize(gitRepoSchema))
	request, err = http.NewRequest("POST", r.serverEndpoint+"/clone/"+orgName, bytes.NewBuffer(payload))
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
	return resp.StatusCode == 200
}

func (r *RepoManagerApiImpl) DeleteRepo(gitRepoSchema git.GitRepoSchema) bool {
	var err error
	var payload []byte
	var request *http.Request
	payload = []byte(serializer.Serialize(gitRepoSchema))
	request, err = http.NewRequest("POST", r.serverEndpoint+"/delete", bytes.NewBuffer(payload))
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
	return resp.StatusCode == 200
}

func (r *RepoManagerApiImpl) SyncRepo(gitRepoSchema git.GitRepoSchema) bool {
	var err error
	var payload []byte
	var request *http.Request
	payload = []byte(serializer.Serialize(gitRepoSchema))
	request, err = http.NewRequest("POST", r.serverEndpoint+"/sync", bytes.NewBuffer(payload))
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
	return resp.StatusCode == 200
}
