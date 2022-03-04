package repomanager

import (
	"github.com/greenopsinc/pipelinereposerver/commandbuilder"
	"github.com/greenopsinc/util/git"
	"os/exec"
	"strings"
	"testing"
)

const (
	GitRepoUrl          = "https://github.com/GreenOpsInc/atlasexamples"
	GitRepoUrl2         = "https://github.com/Saifu0/verificationpipelinetests"
	GitRepoUrlInfected  = "https://github.com/GreenOpsInc/atlasexamples2"
	PathToRoot          = "basic/"
	FileName            = "pipeline.yaml"
	FileNameIncorrect   = "pipeline2.yaml"
	PipelineYamlContent = "name: demo_pipeline\n" +
		"argo_version_lock: true\n" +
		"#cluster_name at the pipeline level means that every step in the pipeline will deploy to cluster_local.\n" +
		"#If cluster_name is also defined at a particular step, it will override the cluster_name set at the pipeline\n" +
		"#(only for that step).\ncluster_name: in-cluster\n" +
		"steps:\n" +
		"- name: deploy_to_dev\n" +
		"  #This is the path of the ArgoCD Application file\n" +
		"  application_path: testapp_dev.yml\n" +
		"  tests:\n" +
		"  - path: verifyendpoints.sh\n" +
		"    type: inject\n" +
		"    image: curlimages/curl:latest\n" +
		"    commands: [sh, -c, ./verifyendpoints.sh]\n" +
		"    before: false\n" +
		"    variables:\n" +
		"      SERVICE_INTERNAL_URL: testapp.dev.svc.cluster.local\n" +
		"- name: deploy_to_int\n" +
		"  application_path: testapp_int.yml\n" +
		"  tests:\n" +
		"  - path: verifyendpoints.sh\n" +
		"    type: inject\n" +
		"    image: curlimages/curl:latest\n" +
		"    commands: [sh, -c, ./verifyendpoints.sh]\n" +
		"    before: false\n" +
		"    variables:\n" +
		"      SERVICE_INTERNAL_URL: testapp.int.svc.cluster.local\n" +
		"  #The schema is DAG based, allowing both linear and complex pipelines\n" +
		"  dependencies:\n" +
		"  - deploy_to_dev\n"
	ExpectedGitCommit = "0c586f73b3094bb4f33942809e2f1fa728467d80"
)

var repoManager RepoManager

func init() {
	repoManager = &RepoManagerImpl{gitRepos: make(map[string]git.GitRepoSchema)}
}

func TestCloneSuccess(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}

	out, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}
}

// when providing an incorrect gitRepoUrl for cloning, cloning is not failing, it is just keep trying to clone. No timeout here.
//func TestCloneShouldFail(t *testing.T) {
//	gitRepoSchema := git.GitRepoSchema{
//		GitRepo:    GitRepoUrlInfected,
//		PathToRoot: PathToRoot,
//		GitCred:    &git.GitCredOpen{},
//	}
//	isCloned := repoManager.Clone(gitRepoSchema)
//	if isCloned {
//		t.Fatalf("Cloning repo %s should have fail.", gitRepoSchema.GitRepo)
//	}
//
//	_, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
//	if ok {
//		t.Fatalf("GitRepo with key %s should not have found in gitRepos registry", getGitRepoKey(gitRepoSchema))
//	}
//}

func TestGetYamlFileContentsSuccess(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	out, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}

	yaml := repoManager.GetYamlFileContents(FileName, gitRepoSchema)
	if yaml != PipelineYamlContent {
		t.Fatalf("Repo %s, PathToRoot %s, file %s content is different than expected.", gitRepoSchema.GetGitRepo(), gitRepoSchema.GetPathToRoot(), FileName)
	}
}

func TestGetYamlFileContentsShouldFail(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	out, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}

	yaml := repoManager.GetYamlFileContents(FileNameIncorrect, gitRepoSchema)
	if yaml == PipelineYamlContent {
		t.Fatalf("Repo %s, PathToRoot %s, file %s content is same as expected.", gitRepoSchema.GetGitRepo(), gitRepoSchema.GetPathToRoot(), FileName)
	}
}

func TestGetCurrentCommitSuccess(t *testing.T) {
	commit := repoManager.GetCurrentCommit(GitRepoUrl)
	if commit != ExpectedGitCommit {
		t.Fatalf("Current commit %s is not same as expected %s.", commit, ExpectedGitCommit)
	}
}

func TestGetCurrentCommitShouldFail(t *testing.T) {
	commit := repoManager.GetCurrentCommit(GitRepoUrlInfected)
	if commit != "" {
		t.Fatalf("Current commit should be an empty string but go %s.", commit)
	}
}

func TestCloneAndSyncSuccess(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	out, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}

	gitCommit := repoManager.GetCurrentCommit(gitRepoSchema.GetGitRepo())
	syncedCommit := repoManager.Sync(gitRepoSchema)
	if syncedCommit != gitCommit {
		t.Fatalf("Syncing repo %s failed. expected commit hash %s, got %s", gitRepoSchema.GitRepo, gitCommit, syncedCommit)
	}
	out, ok = repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}
}

func TestCloneAndSyncShouldFail(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	out, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}

	gitCommit := repoManager.GetCurrentCommit(GitRepoUrl2)
	syncedCommit := repoManager.Sync(gitRepoSchema)
	if syncedCommit == gitCommit {
		t.Fatalf("Syncing repo %s should failed. Got same commit hash %s", gitRepoSchema.GitRepo, gitCommit)
	}
	out, ok = repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}
}

func TestCloneAndDeleteSuccess(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	out, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
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

func TestCloneAndDeleteShouldFail(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	out, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}

	gitRepoSchema.GitRepo = GitRepoUrlInfected
	isDeleted := repoManager.Delete(gitRepoSchema)
	if !isDeleted {
		t.Fatalf("Deleting repo %s should failed but successful", gitRepoSchema.GitRepo)
	}
	_, ok = repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if ok {
		t.Fatalf("GitRepo with key %s should not have found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
}

func TestCloneAndUpdateSuccess(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl2,
		PathToRoot: "",
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	out, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}

	isUpdated := repoManager.Update(gitRepoSchema)
	if !isUpdated {
		t.Fatalf("Upating repo %s failed.", gitRepoSchema.GitRepo)
	}
	out, ok = repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}
}

func TestCloneAndUpdateShouldFail(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	out, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
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

func TestResetToVersionSuccess(t *testing.T) {
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
	out, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}

	isResetToVersion := repoManager.ResetToVersion(gitCommit, gitRepoSchema)
	if !isResetToVersion {
		t.Fatalf("Reseting to %s failed for repo %s", gitCommit, gitRepoSchema.GetGitRepo())
	}
}

func TestResetToVersionShouldFail(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	out, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}

	gitCommit := "incorrectCommitHash"
	isResetToVersion := repoManager.ResetToVersion(gitCommit, gitRepoSchema)
	if isResetToVersion {
		t.Fatalf("Reseting to %s should fail for repo %s", gitCommit, gitRepoSchema.GetGitRepo())
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
	out, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}

	for _, commit := range commits {
		isResetToVersion := repoManager.ResetToVersion(commit, gitRepoSchema)
		if !isResetToVersion {
			t.Fatalf("Reseting to %s failed for repo %s", commit, gitRepoSchema.GetGitRepo())
		}
	}
}

func TestContainsGitRepoSchemaSuccess(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	out, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}

	isContain := repoManager.ContainsGitRepoSchema(gitRepoSchema)
	if !isContain {
		t.Fatalf("%s repo not exist in gitRepo registry", gitRepoSchema.GetGitRepo())
	}
}

func TestContainsGitRepoSchemaShouldFail(t *testing.T) {
	gitRepoSchema := git.GitRepoSchema{
		GitRepo:    GitRepoUrl,
		PathToRoot: PathToRoot,
		GitCred:    &git.GitCredOpen{},
	}
	isCloned := repoManager.Clone(gitRepoSchema)
	if !isCloned {
		t.Fatalf("Cloning repo %s failed.", gitRepoSchema.GitRepo)
	}
	out, ok := repoManager.GetGitRepos()[getGitRepoKey(gitRepoSchema)]
	if !ok {
		t.Fatalf("GitRepo with key %s not found in gitRepos registry", getGitRepoKey(gitRepoSchema))
	}
	if out != gitRepoSchema {
		t.Fatalf("local gitRepoSchema differs from gitRepoSchema from gitRepos registry")
	}

	gitRepoSchema.GitRepo = GitRepoUrl2
	isContain := repoManager.ContainsGitRepoSchema(gitRepoSchema)
	if isContain {
		t.Fatalf("%s repo should not exist in gitRepo registry", gitRepoSchema.GetGitRepo())
	}
}
