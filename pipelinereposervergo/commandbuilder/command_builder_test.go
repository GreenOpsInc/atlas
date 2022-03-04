package commandbuilder

import (
	"github.com/greenopsinc/util/git"
	"os"
	"os/exec"
	"testing"
)

var commandBuilder *CommandBuilder

const (
	GitRepoUrl           = "https://github.com/GreenOpsInc/atlasexamples"
	PathToRoot           = "basic/"
	test                 = "test"
	atlasexample         = "atlasexample"
	commitHash           = "994a6c9f2b25aa3fec84f823366d53f61898bd71"
	expectedMultipleLogs = "commit 0c586f73b3094bb4f33942809e2f1fa728467d80\n" +
		"Author: mihirpandya-greenops <84353489+mihirpandya-greenops@users.noreply.github.com>\n" +
		"Date:   Mon Feb 14 17:16:02 2022 -0800\n\n" +
		"    Update verifyendpoints.sh\n\n" +
		"commit 994a6c9f2b25aa3fec84f823366d53f61898bd71\n" +
		"Author: mihirpandya-greenops <mihir@greenops.io>\n" +
		"Date:   Fri Jan 28 13:55:53 2022 -0800\n\n" +
		"    updating rollouts example\n"
)

var (
	dir = os.Getenv("HOME") + "/go/src/github.com/greenops/atlas/pipelinereposervergo/commandbuilder"
)

func init() {
	commandBuilder = New()
}

func TestMkdir(t *testing.T) {
	var cmd = commandBuilder.Mkdir(test).Build()
	output, err := exec.Command("/bin/bash", "-c", cmd).Output()
	if err != nil {
		deleteCmd(test)
		t.Fatalf("unable to create a new directory, error: %s-%s", err.Error(), output)
	}

	output, err = exec.Command("/bin/bash", "-c", cmd).Output()
	if err == nil {
		t.Fatalf("created new directory again, error: %s-%s", err.Error(), output)
	}
	deleteCmd(test)
}

func TestCd(t *testing.T) {
	cmd := commandBuilder.Mkdir(test).
		Cd(test).
		Build()
	process := exec.Command("/bin/bash", "-c", cmd)
	process.Dir = dir
	output, err := process.Output()
	if err != nil {
		deleteCmd(test)
		t.Errorf("Changing directory to test was not successful. Error: %s - %s", err, output)
	}
	deleteCmd(test)
}

func TestGitClone(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	cmd := commandBuilder.GitClone(gitRepoSchema).
		DeleteFolder(gitRepoSchema).
		Build()
	process := exec.Command("/bin/bash", "-c", cmd)
	process.Dir = dir
	output, err := process.Output()
	if err != nil {
		deleteCmd(atlasexample)
		t.Errorf("Cloning repo %s was not successful. Error: %s - %s", gitRepoSchema.GitRepo, err, output)
	}
}

func TestGitPull(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	cmd := commandBuilder.GitClone(gitRepoSchema).
		Cd(GetFolderName(gitRepoSchema.GetGitRepo())).
		GitPull(gitRepoSchema).
		Cd("..").
		DeleteFolder(gitRepoSchema).
		Build()
	process := exec.Command("/bin/bash", "-c", cmd)
	process.Dir = dir
	output, err := process.Output()
	if err != nil {
		deleteCmd(atlasexample)
		t.Errorf("Pulling repo %s was not successful. Error: %s - %s", gitRepoSchema.GitRepo, err, output)
	}
}

func TestGitCheckout(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	cmd := commandBuilder.GitClone(gitRepoSchema).
		Cd(GetFolderName(gitRepoSchema.GetGitRepo())).
		GitCheckout(commitHash).
		Cd("..").
		DeleteFolder(gitRepoSchema).
		Build()
	process := exec.Command("/bin/bash", "-c", cmd)
	process.Dir = dir
	output, err := process.Output()
	if err != nil {
		deleteCmd(atlasexample)
		t.Errorf("Checking out to %s at repo %s was not successful. Error: %s - %s", commitHash, gitRepoSchema.GitRepo, err, output)
	}
}

func TestGitLog(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	cmd := commandBuilder.GitClone(gitRepoSchema).
		Cd(GetFolderName(gitRepoSchema.GetGitRepo())).
		GitLog(1, true).
		Cd("..").
		DeleteFolder(gitRepoSchema).
		Build()
	process := exec.Command("/bin/bash", "-c", cmd)
	process.Dir = dir
	output, err := process.Output()
	if err != nil {
		deleteCmd(atlasexample)
		t.Errorf("Testing Logs for repo %s was not successful. Error: %s - %s", gitRepoSchema.GitRepo, err, output)
	}

	if string(output) != "0c586f73b3094bb4f33942809e2f1fa728467d80" {
		deleteCmd(atlasexample)
		t.Errorf("Log checking failed, expected %s, got %s", "0c586f73b3094bb4f33942809e2f1fa728467d80", string(output))
	}
}

func TestComplexCommands(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	cmd := commandBuilder.Mkdir(test).
		Cd(test).
		GitClone(gitRepoSchema).
		Cd(GetFolderName(gitRepoSchema.GetGitRepo())).
		GitLog(2, false).
		Build()
	process := exec.Command("/bin/bash", "-c", cmd)
	process.Dir = dir
	output, err := process.Output()
	if err != nil {
		deleteCmd(test)
		t.Errorf("Test on complex commands was not successful. Error: %s - %s", err, output)
	}
	if string(output) != expectedMultipleLogs {
		deleteCmd(test)
		t.Errorf("Logs checking failed, expected %s, got %s", expectedMultipleLogs, string(output))
	}
	deleteCmd(test)
}

func deleteCmd(dir string) {
	_ = exec.Command("rm", "-rf", dir).Run()
}
