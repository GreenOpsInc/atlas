package reposerver

import (
	"strings"

	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/httpclient"
	"github.com/greenopsinc/util/tlsmanager"
)

const (
	rootDataExtension  = "data"
	rootRepoExtension  = "repo"
	getFileExtension   = "file"
	getCommitExtension = "version"
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

func (r repoManagerAPI) GetFileFromRepo(getFileRequest *git.GetFileRequest, orgName, teamName string) (string, error) {
	panic("implement me")
}

func (r repoManagerAPI) ResetRepoVersion(gitCommit string, gitRepoSchemaInfo *git.GitRepoSchemaInfo, orgName, teamName string) error {
	panic("implement me")
}
