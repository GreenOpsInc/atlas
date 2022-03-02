package commandbuilder

import (
	"github.com/greenopsinc/pipelinereposerver/api/util"
	"github.com/greenopsinc/util/git"
	"strconv"
	"strings"
)

type CommandBuilder struct {
	commands []string
}

func New() *CommandBuilder {
	return &CommandBuilder{
		commands: make([]string, 0),
	}
}

func (c *CommandBuilder) Mkdir(directory string) *CommandBuilder {
	c.commands = append(c.commands, "mkdir "+directory)
	return c
}

func (c *CommandBuilder) Cd(directory string) *CommandBuilder {
	c.commands = append(c.commands, "cd "+directory)
	return c
}

func (c *CommandBuilder) GitClone(gitRepoSchema git.GitRepoSchema) *CommandBuilder {
	newCommand := make([]string, 0)
	newCommand = append(newCommand, "git clone")
	newCommand = append(newCommand, util.GetGitCredAccessibleFromGitCred(gitRepoSchema.GetGitCred()).ConvertGitCredToString(gitRepoSchema.GetGitRepo()))
	c.commands = append(c.commands, strings.Join(newCommand, " "))
	return c
}

func (c *CommandBuilder) GitPull(gitRepoSchema git.GitRepoSchema) *CommandBuilder {
	newCommand := make([]string, 0)
	newCommand = append(newCommand, "git pull")
	newCommand = append(newCommand, util.GetGitCredAccessibleFromGitCred(gitRepoSchema.GetGitCred()).ConvertGitCredToString(gitRepoSchema.GetGitRepo()))
	c.commands = append(c.commands, strings.Join(newCommand, " "))
	return c
}

func (c *CommandBuilder) GitCheckout(commitHash string) *CommandBuilder {
	newCommand := make([]string, 0)
	newCommand = append(newCommand, "git checkout")
	newCommand = append(newCommand, commitHash)
	c.commands = append(c.commands, strings.Join(newCommand, " "))
	return c
}

func (c *CommandBuilder) GitLog(logCount int, justCommits bool) *CommandBuilder {
	if justCommits {
		c.commands = append(c.commands, "git log -n "+strconv.Itoa(logCount)+" --pretty=format:\"%H\"")
	} else {
		c.commands = append(c.commands, "git log -n "+strconv.Itoa(logCount))
	}
	return c
}

func (c *CommandBuilder) DeleteFolder(gitRepoSchema git.GitRepoSchema) *CommandBuilder {
	newCommand := make([]string, 0)
	newCommand = append(newCommand, "rm -rf")
	newCommand = append(newCommand, GetFolderName(gitRepoSchema.GetGitRepo()))
	c.commands = append(c.commands, strings.Join(newCommand, " "))
	return c
}

func (c *CommandBuilder) Build() string {
	return strings.Join(c.commands, "; ")
}

func GetFolderName(gitRepo string) string {

	splitLink := strings.Split(gitRepo, "/")
	idx := len(splitLink) - 1
	for idx >= 0 {
		if splitLink[idx] == "" {
			idx--
		} else {
			break
		}
	}
	return strings.Replace(splitLink[idx], ".git", "", -1)
}
