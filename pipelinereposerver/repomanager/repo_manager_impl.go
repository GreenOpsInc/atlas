package repomanager

import (
	"github.com/greenopsinc/pipelinereposerver/commandbuilder"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/kubernetesclient"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	DefaultOrg string = "org"
	directory  string = "tmp"
)

type RepoManager interface {
	GetOrgName() string
	GetGitRepos() map[string]git.GitRepoSchema
	Clone(gitRepoSchema git.GitRepoSchema) bool
	Update(gitRepoSchema git.GitRepoSchema) bool
	Delete(gitRepoSchema git.GitRepoSchema) bool
	GetYamlFileContents(filename string, gitRepoSchema git.GitRepoSchema) string
	Sync(gitRepoSchema git.GitRepoSchema) string
	ResetToVersion(gitCommit string, gitRepoSchema git.GitRepoSchema) bool
	ContainsGitRepoSchema(gitRepoSchema git.GitRepoSchema) bool
	GetCurrentCommit(gitRepoURL string) string
}

type RepoManagerImpl struct {
	orgName  string
	gitRepos map[string]git.GitRepoSchema
}

func New(dbOperator db.DbOperator, kubernetesClient kubernetesclient.KubernetesClient, orgName string) RepoManager {
	gitReposMap := make(map[string]git.GitRepoSchema)
	repoManager := &RepoManagerImpl{
		orgName:  orgName,
		gitRepos: gitReposMap,
	}
	if !repoManager.setupRepoCache(dbOperator, kubernetesClient) {
		log.Fatal("The org's repos could not be cloned correctly. Please restart to try again.")
	}
	return repoManager
}

func (r *RepoManagerImpl) GetOrgName() string {
	return r.orgName
}

func (r *RepoManagerImpl) GetGitRepos() map[string]git.GitRepoSchema {
	return r.gitRepos
}

func (r *RepoManagerImpl) Clone(gitRepoSchema git.GitRepoSchema) bool {
	if r.ContainsGitRepoSchema(gitRepoSchema) {
		return true
	}
	command := commandbuilder.New().
		GitClone(gitRepoSchema).
		Cd(commandbuilder.GetFolderName(gitRepoSchema.GetGitRepo())).
		GitLog(1, true).
		Build()

	process := exec.Command("/bin/bash", "-c", command)
	process.Dir = r.orgName + "/" + directory
	output, err := process.Output()
	if err != nil {
		log.Printf("Cloning repo %s was not successful. Error: %s - %s\nCleaning up...", gitRepoSchema.GetGitRepo(), err, output)
		r.Delete(gitRepoSchema)
		return false
	} else {
		log.Printf("Cloning repo %s was successful.", gitRepoSchema.GetGitRepo())
		r.gitRepos[getGitRepoKey(gitRepoSchema)] = gitRepoSchema
		return true
	}
}

func (r *RepoManagerImpl) Update(gitRepoSchema git.GitRepoSchema) bool {
	oldGitRepo, ok := r.gitRepos[getGitRepoKey(gitRepoSchema)]
	if !ok {
		return false
	}
	delete(r.gitRepos, getGitRepoKey(gitRepoSchema))
	r.gitRepos[getGitRepoKey(gitRepoSchema)] = gitRepoSchema
	if r.Sync(gitRepoSchema) != "" {
		return true
	} else {
		delete(r.gitRepos, getGitRepoKey(gitRepoSchema))
		r.gitRepos[getGitRepoKey(oldGitRepo)] = oldGitRepo
		return false
	}
}

func (r *RepoManagerImpl) Delete(gitRepoSchema git.GitRepoSchema) bool {
	command := commandbuilder.New().
		DeleteFolder(gitRepoSchema).
		Build()

	process := exec.Command("/bin/bash", "-c", command)
	process.Dir = r.orgName + "/" + directory
	output, err := process.Output()
	if err != nil {
		log.Printf("An error was thrown when attempting to delete the repo %s - %s", err, output)
		return false
	} else {
		delete(r.gitRepos, getGitRepoKey(gitRepoSchema))
		return true
	}
}

func (r *RepoManagerImpl) GetYamlFileContents(filename string, gitRepoSchema git.GitRepoSchema) string {
	if _, ok := r.gitRepos[getGitRepoKey(gitRepoSchema)]; !ok {
		log.Printf("Repo does not exist in manager")
		return ""
	}
	truncatedFilePath := strings.Trim(filename, "/")
	pathToRoot := strings.Trim(gitRepoSchema.GetPathToRoot(), "/")
	var fullPath string
	if pathToRoot == "" {
		fullPath = strings.Join([]string{r.orgName, directory, commandbuilder.GetFolderName(gitRepoSchema.GetGitRepo()), truncatedFilePath}, "/")
	} else {
		fullPath = strings.Join([]string{r.orgName, directory, commandbuilder.GetFolderName(gitRepoSchema.GetGitRepo()), pathToRoot, truncatedFilePath}, "/")
	}
	dat, err := os.ReadFile(fullPath)
	if err != nil {
		log.Printf("Error reading file: %s", err)
		return ""
	}
	return string(dat)
}

func (r *RepoManagerImpl) Sync(gitRepoSchema git.GitRepoSchema) string {
	cachedGitRepoSchema, ok := r.gitRepos[getGitRepoKey(gitRepoSchema)]
	if !ok {
		log.Printf("Repo does not exist in manager")
		return ""
	}

	command := commandbuilder.New().GitCheckout("main").GitPull(cachedGitRepoSchema).Build()
	process := exec.Command("/bin/bash", "-c", command)
	process.Dir = r.orgName + "/" + directory + "/" + commandbuilder.GetFolderName(cachedGitRepoSchema.GetGitRepo())
	output, err := process.Output()
	if err != nil {
		log.Printf("Pulling repo %s was not successful: %s - %s", cachedGitRepoSchema.GetGitRepo(), err, output)
		return ""
	} else {
		log.Printf("Pulling repo %s was successful.", cachedGitRepoSchema.GetGitRepo())
		commitHash := r.GetCurrentCommit(cachedGitRepoSchema.GetGitRepo())
		if commitHash == "" {
			return ""
		}
		return commitHash
	}
}

func (r *RepoManagerImpl) ResetToVersion(gitCommit string, gitRepoSchema git.GitRepoSchema) bool {
	cachedGitRepoSchema, ok := r.gitRepos[getGitRepoKey(gitRepoSchema)]
	if !ok {
		log.Printf("Repo does not exist in manager")
		return false
	}

	command := commandbuilder.New().GitCheckout(gitCommit).Build()
	process := exec.Command("/bin/bash", "-c", command)
	process.Dir = r.orgName + "/" + directory + "/" + commandbuilder.GetFolderName(cachedGitRepoSchema.GetGitRepo())
	output, err := process.Output()
	if err != nil {
		log.Printf("Updating repo version %s was not successful: %s - %s", cachedGitRepoSchema.GetGitRepo(), err, output)
		return false
	} else {
		log.Printf("Updating repo version %s was successful", cachedGitRepoSchema.GetGitRepo())
		return true
	}
}

func (r *RepoManagerImpl) GetCurrentCommit(gitRepoUrl string) string {
	command := commandbuilder.New().GitLog(1, true).Build()
	process := exec.Command("/bin/bash", "-c", command)
	process.Dir = r.orgName + "/" + directory + "/" + commandbuilder.GetFolderName(gitRepoUrl)
	output, err := process.Output()
	if err != nil {
		log.Printf("Fetching commit was not successful: %s - %s", err, output)
		return ""
	} else {
		log.Printf("Fetching commit was successful")
		return strings.Split(string(output), "\n")[0]
	}
}

func (r *RepoManagerImpl) ContainsGitRepoSchema(gitRepoSchema git.GitRepoSchema) bool {
	if _, ok := r.gitRepos[getGitRepoKey(gitRepoSchema)]; ok {
		return true
	}
	return false
}

func getGitRepoKey(gitRepoSchema git.GitRepoSchema) string {
	return gitRepoSchema.GetGitRepo() + gitRepoSchema.GetPathToRoot()
}

func strip(str string, delimiter string) string {
	beg := 0
	end := len(str) - 1
	for beg < len(str) && str[beg:beg+1] == delimiter {
		beg++
	}
	for end >= 0 && str[end:end+1] == delimiter {
		end--
	}
	if end < beg {
		return ""
	} else {
		return str[beg : end+1]
	}
}

func (r *RepoManagerImpl) setupRepoCache(dbOperator db.DbOperator, kubernetesClient kubernetesclient.KubernetesClient) bool {
	defer dbOperator.Close()
	command := commandbuilder.New().
		Mkdir(r.orgName).
		Cd(r.orgName).
		Mkdir(directory).
		Cd(directory).Build()

	output, err := exec.Command("/bin/bash", "-c", command).Output()
	if err != nil {
		log.Printf("Errors: %s - %s", err, output)
		return false
	}

	dbClient := dbOperator.GetClient()
	defer dbClient.Close()
	listOfTeams := dbClient.FetchStringList(db.MakeDbListOfTeamsKey(r.orgName))
	if len(listOfTeams) == 0 {
		log.Printf("No teams in org %s", r.orgName)
		return true
	}
	log.Printf("Fetched all teams and cloning pipeline repos for org %s", r.orgName)
	for _, teamName := range listOfTeams {
		teamSchema := dbClient.FetchTeamSchema(db.MakeDbTeamKey(r.orgName, teamName))
		if teamSchema.GetOrgName() == "" {
			//means team is nil
			log.Printf("The team %s doesn't exist, so cloning will be skipped", teamName)
			continue
		}
		for _, pipelineSchema := range teamSchema.GetPipelineSchemas() {
			gitRepoSchema := pipelineSchema.GetGitRepoSchema()
			secretName := db.MakeSecretName(teamSchema.GetOrgName(), teamName, pipelineSchema.GetPipelineName())
			gitCred := kubernetesClient.FetchGitCred(secretName)
			if gitCred == nil {
				panic("Could not recover gitCred from kubernetes secrets")
			}
			gitRepoSchema.SetGitCred(gitCred)
			if !r.Clone(gitRepoSchema) {
				return false
			}
		}
		log.Printf("Finished cloning pipeline repos for team %s", teamName)
	}
	return true
}
