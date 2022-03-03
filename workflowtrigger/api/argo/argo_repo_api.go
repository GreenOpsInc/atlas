package argo

import (
	"context"
	"fmt"
	repositorypkg "github.com/argoproj/argo-cd/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/greenopsinc/util/git"
	"log"
	"strings"
)

type ArgoRepoApi interface {
	CreateRepo(gitRepo string, gitCred git.GitCred) error
}

func (a *ArgoApiImpl) CreateRepo(gitRepo string, gitCred git.GitCred) error {
	closer, client, err := a.configuredClient.NewRepoClient()
	if err != nil {
		panic(fmt.Sprintf("cluster client could not be made for Argo: %s", err))
	}
	defer closer.Close()
	var repo *v1alpha1.Repository
	switch gitCred.(type) {
	case *git.GitCredMachineUser:
		repo = &v1alpha1.Repository{
			Repo:     gitRepo,
			Username: gitCred.(*git.GitCredMachineUser).Username,
			Password: gitCred.(*git.GitCredMachineUser).Password,
		}
	case *git.GitCredOpen:
		repo = &v1alpha1.Repository{
			Repo: gitRepo,
		}
	}
	exists, err := a.CheckRepoExists(gitRepo)
	if err != nil {
		return err
	}
	if !exists {
		_, err = client.Create(context.TODO(), &repositorypkg.RepoCreateRequest{
			Repo:   repo,
			Upsert: false,
		})
	}
	if err != nil {
		log.Printf("Creating the repo failed with error: %s", err)
		return err
	}
	return nil
}

func (a *ArgoApiImpl) CheckRepoExists(gitRepo string) (bool, error) {
	closer, client, err := a.configuredClient.NewRepoClient()
	if err != nil {
		panic(fmt.Sprintf("cluster client could not be made for Argo: %s", err))
	}
	defer closer.Close()

	_, err = client.Get(context.TODO(), &repositorypkg.RepoQuery{Repo: gitRepo})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		log.Printf("Error occurred when checking for the existance of the repo in Argo: %s", err)
		return false, err
	}

	return true, nil
}
