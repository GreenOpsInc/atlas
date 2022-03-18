package data

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/greenopsinc/util/pipeline/request"
)

const (
	CustomTask       = "custom"
	InjectTask       = "inject"
	WorkflowTask     = "ArgoWorkflowTask"
	DefaultNamespace = "default"
	WatchTestKey     = "WatchTestKey"
)

type TestData interface {
	GetPayload(testNumber int, testConfig string) interface{}
}

type CustomJobTest struct {
	Path                    string        `json:"path"`
	ExecuteBeforeDeployment bool          `json:"executeBeforeDeployment"`
	Variables               []interface{} `json:"variables"`
}

func (t *CustomJobTest) GetPayload(_ int, testConfig string) interface{} {
	return &request.KubernetesCreationRequest{
		Type:          CustomTask,
		ConfigPayload: testConfig,
		Variables:     t.Variables,
	}
}

type InjectScriptTest struct {
	Path                    string        `json:"path"`
	Image                   string        `json:"image"`
	Namespace               string        `json:"namespace"`
	Commands                []string      `json:"commands"`
	Arguments               []string      `json:"arguments"`
	ExecuteInApplicationPod bool          `json:"executeInApplicationPod"`
	ExecuteBeforeDeployment bool          `json:"executeBeforeDeployment"`
	Variables               []interface{} `json:"variables"`
}

func (t *InjectScriptTest) GetPayload(testNumber int, testConfig string) (interface{}, error) {
	filename := filepath.Base(t.Path)

	var imageName string
	if t.Image != "" {
		imageName = t.Image
	} else {
		return nil, errors.New("image name for InjectScriptTest should not be empty")
	}

	jobNamespace := DefaultNamespace
	if t.Namespace != "" {
		jobNamespace = t.Namespace
	}

	testKey := strings.Join(
		[]string{
			fmt.Sprintf("%d", testNumber),
			uuid.New().String(),
		},
		"-",
	)

	return &request.KubernetesCreationRequest{
		Type:           InjectTask,
		Kind:           "Job",
		ObjectName:     testKey,
		Namespace:      jobNamespace,
		ImageName:      imageName,
		Command:        t.Commands,
		Args:           t.Arguments,
		ConfigPayload:  "",
		VolumeFilename: filename,
		VolumePayload:  testConfig,
		Variables:      t.Variables,
	}, nil
}

type ArgoWorkflowTask struct {
	Path                    string        `json:"path"`
	ExecuteBeforeDeployment bool          `json:"execute_before_deployment"`
	Variables               []interface{} `json:"variables"`
}

func (t *ArgoWorkflowTask) GetPayload(_ int, testConfig string) interface{} {
	return &request.KubernetesCreationRequest{
		Type:          WorkflowTask,
		ConfigPayload: testConfig,
		Variables:     t.Variables,
	}
}
