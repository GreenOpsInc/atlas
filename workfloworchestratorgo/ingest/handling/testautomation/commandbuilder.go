package testautomation

import "strings"

type CommandBuilder interface {
	CreateFile(fileName, fileContents string) CommandBuilder
	Compile(fileName string) CommandBuilder
	LS() CommandBuilder
	CAT(fileName string) CommandBuilder
	ExecuteFileContents(fileContents string) CommandBuilder
	ExecuteExistingFile(fileName string) CommandBuilder
	Build() []string
}

type commandBuilder struct {
	commands []string
}

func NewCommandBuilder() CommandBuilder {
	return &commandBuilder{commands: []string{}}
}

func (c *commandBuilder) CreateFile(fileName, fileContents string) CommandBuilder {
	cmd := "echo " + "'" + fileContents + "'" + " > " + fileName
	c.commands = append(c.commands, cmd)
	return c
}

func (c *commandBuilder) Compile(fileName string) CommandBuilder {
	cmd := "chmod +x " + fileName
	c.commands = append(c.commands, cmd)
	return c
}

func (c *commandBuilder) LS() CommandBuilder {
	c.commands = append(c.commands, "ls")
	return c
}

func (c *commandBuilder) CAT(fileName string) CommandBuilder {
	cmd := "cat " + fileName
	c.commands = append(c.commands, cmd)
	return c
}

func (c *commandBuilder) ExecuteFileContents(fileContents string) CommandBuilder {
	c.commands = append(c.commands, fileContents)
	return c
}

func (c *commandBuilder) ExecuteExistingFile(fileName string) CommandBuilder {
	c.commands = append(c.commands, "./"+fileName)
	return c
}

func (c *commandBuilder) Build() []string {
	return []string{strings.Join(c.commands, "\n")}
}
