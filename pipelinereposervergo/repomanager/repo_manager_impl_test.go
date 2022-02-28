package repomanager

import (
	"github.com/greenopsinc/pipelinereposerver/commandbuilder"
	"github.com/greenopsinc/util/git"
	"os/exec"
	"strings"
	"testing"
)

const (
	orgName            = "org"
	GitRepoUrl         = "https://github.com/GreenOpsInc/atlasexamples"
	GitRepoUrlInfected = "https://github.com/GreenOpsInc/atlasexamples2"
	PathToRoot         = "basic/"
)

var repoManager RepoManager

func init() {
	repoManager = &RepoManagerImpl{gitRepos: make(map[string]git.GitRepoSchema)}
}

func TestCloneAndDelete(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	_, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}

	isDeleted := repoManager.Delete(gitRepoSchema)
	if !isDeleted {
		t.Fatalf("Deleting repo %s failed.", gitRepoSchema.GitRepo)
	}
	_, ok = repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if ok {
		t.Fatalf("GitRepo with key %s is still found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
}

func TestCloneAndUpdateSuccess(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	_, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}

	isUpdated := repoManager.Update(gitRepoSchema)
	if !isUpdated {
		t.Fatalf("Upating repo %s failed.", gitRepoSchema.GitRepo)
	}
	_, ok = repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
}

func TestCloneAndUpdateFailed(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	_, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}

	gitRepoSchema.GitRepo = GitRepoUrlInfected
	isUpdated := repoManager.Update(gitRepoSchema)
	if isUpdated {
		t.Fatalf("Update Should fail for repo %s.", gitRepoSchema.GitRepo)
	}
	_, ok = repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if ok {
		t.Fatalf("GitRepo with key %s should not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
}

func TestResetToVersion(t *testing.T) {
	gitCommit := repoManager.GetCurrentCommit(GitRepoUrl)
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	_, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}

	isResetToVersion := repoManager.ResetToVersion(gitCommit, gitRepoSchema)
	if !isResetToVersion {
		t.Fatalf("Reseting to %s failed for repo %s", gitCommit, gitRepoSchema.GetGitRepo())
	}
}

func TestMultipleResetToVersion(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	cmd := commandbuilder.New().
		GitClone(gitRepoSchema).
		Cd(commandbuilder.GetFolderName(gitRepoSchema.GetGitRepo())).
		GitLog(3, true).
		Cd("..").
		DeleteFolder(gitRepoSchema).
		Build()
	process := exec.Command("/bin/bash", "-c", cmd)
	output, err := process.Output()
	if err != nil {
		repoManager.Delete(gitRepoSchema)
		t.Errorf("Getting commits for repo %s was not successful. Error: %s - %s", gitRepoSchema.GitRepo, err, output)
	}

	commits := strings.Split(string(output), "\n")
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	_, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	for _, commit := range commits {
		isResetToVersion := repoManager.ResetToVersion(commit, gitRepoSchema)
		if !isResetToVersion {
			t.Fatalf("Reseting to %s failed for repo %s", commit, gitRepoSchema.GetGitRepo())
		}
	}
}
