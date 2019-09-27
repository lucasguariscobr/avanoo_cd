package utils

import (
	"fmt"
	"time"

	"github.com/stretchr/testify/mock"
)

var ExecCommand = processCommand

type Commandable interface {
	String() string
	CombinedOutput() ([]byte, error)
}

type BuildCommand struct {
	Commandable
}

type DeployCommand struct {
	Commandable
}

func processCommand(c Commandable) ([]byte, error) {
	return c.CombinedOutput()
}

type TestCommand struct {
	mock.Mock
}

func (c *TestCommand) CombinedOutput() ([]byte, error) {
	args := c.Called()
	var output []byte
	if args.Get(0) != nil {
		output = args.Get(0).([]byte)
	}
	return output, args.Error(1)
}

type CommandMockReturn struct {
	ProcessTime int
	Byte        []byte
	Error       error
}

func MockCommandDefault() {
	MockCommand(&CommandMockReturn{}, &CommandMockReturn{})
}
func MockCommand(build *CommandMockReturn, deploy *CommandMockReturn) {
	testObj := new(TestCommand)
	ExecCommand = func(c Commandable) ([]byte, error) {
		switch c.(type) {
		case BuildCommand:
			fmt.Println("BUILD COMMAND")
			time.Sleep(time.Duration(build.ProcessTime) * time.Second)
			testObj.On("CombinedOutput").Return(build.Byte, build.Error)
		case DeployCommand:
			fmt.Println("DEPLOY COMMAND")
			time.Sleep(time.Duration(build.ProcessTime) * time.Second)
			testObj.On("CombinedOutput").Return(deploy.Byte, deploy.Error)
		}
		return testObj.CombinedOutput()
	}
}
