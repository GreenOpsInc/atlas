package reposerver

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/greenopsinc/util/serializer"
	"github.com/greenopsinc/workfloworchestrator/ingest/apiclient/util"

	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/httpclient"
	"github.com/greenopsinc/util/tlsmanager"
)

const (
	rootDataExtension      = "data"
	rootRepoExtension      = "repo"
	getFileExtension       = "file"
	getCommitExtension     = "version"
	changeVersionExtension = "resetToVersion"
)

type RepoManagerAPI interface {
	GetFileFromRepo(getFileRequest *git.GetFileRequest, orgName, teamName string) (string, error)
	ResetRepoVersion(gitCommit string, gitRepoSchemaInfo *git.GitRepoSchemaInfo, orgName, teamName string) error
}

type repoManagerAPI struct {
	serverRepoEndpoint string
	serverDataEndpoint string
	httpClient         httpclient.HttpClient
}

func NewRepoManagerAPI(tm tlsmanager.Manager, repoServerEndpoint string) (RepoManagerAPI, error) {
	var repoEndpoint, dataEndpoint string
	if strings.HasSuffix(repoServerEndpoint, "/") {
		repoEndpoint = repoServerEndpoint + rootRepoExtension
		dataEndpoint = repoServerEndpoint + rootDataExtension
	} else {
		repoEndpoint = repoServerEndpoint + "/" + rootRepoExtension
		dataEndpoint = repoServerEndpoint + "/" + rootDataExtension
	}

	httpClient, err := httpclient.New(tlsmanager.ClientRepoServer, tm)
	if err != nil {
		return nil, err
	}
	return &repoManagerAPI{
		serverRepoEndpoint: repoEndpoint,
		serverDataEndpoint: dataEndpoint,
		httpClient:         httpClient,
	}, nil
}

func (r *repoManagerAPI) GetFileFromRepo(getFileRequest *git.GetFileRequest, orgName, teamName string) (string, error) {
	url := fmt.Sprintf("%s/%s/%s/%s", r.serverDataEndpoint, getFileExtension, orgName, teamName)
	var request *http.Request
	payload := []byte(serializer.Serialize(getFileRequest))
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	res, err := r.httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if err = util.CheckResponseStatus(res); err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(res.Body)
	return buf.String(), nil
}

func (r *repoManagerAPI) ResetRepoVersion(gitCommit string, gitRepoSchemaInfo *git.GitRepoSchemaInfo, orgName, teamName string) error {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", r.serverDataEndpoint, changeVersionExtension, orgName, teamName, gitCommit)
	var request *http.Request
	payload := []byte(serializer.Serialize(gitRepoSchemaInfo))
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	res, err := r.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if err = util.CheckResponseStatus(res); err != nil {
		return err
	}
	return nil
}
