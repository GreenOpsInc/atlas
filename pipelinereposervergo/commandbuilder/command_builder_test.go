package commandbuilder

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

var commandBuilder *CommandBuilder

func init() {
	commandBuilder = New()
}

func TestMkdir(t *testing.T) {
	var c = commandBuilder.Mkdir("test")
	for _, command := range c.commands {
		args := strings.Fields(command)
		err := exec.Command(args[0], args[1:]...).Run()
		if err != nil {
			t.Fatalf("unable to create a new directory, error: %s", err.Error())
		}

		err = exec.Command(args[0], args[1:]...).Run()
		if err == nil {
			t.Fatalf("created new directory again, error: %s", err.Error())
		}
	}
	_ = exec.Command("rm", "-rf", "test").Run()
}

func TestCd(t *testing.T) {
	var c = commandBuilder.Mkdir("test")
	c = commandBuilder.Cd("test")
	for _, command := range c.commands {
		fmt.Println(command)
		args := strings.Fields(command)
		err := exec.Command(args[0], args[1:]...).Run()
		if err != nil {
			t.Fatalf(err.Error())
		}
	}
	_ = exec.Command("rm", "-rf", "test").Run()
}
