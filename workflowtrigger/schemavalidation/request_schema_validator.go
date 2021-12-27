package schemavalidation

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"greenops.io/workflowtrigger/api/argoauthenticator"
	"greenops.io/workflowtrigger/api/reposerver"
)

const (
	defaultProject string = "default"
)

type PipelineData struct {
	ClusterName string `yaml:"cluster_name"`
	Steps       []StepData
}

type StepData struct {
	Name            string `yaml:"name"`
	ClusterName     string `yaml:"cluster_name"`
	ApplicationPath string `yaml:"application_path"`
}

func (p *PipelineData) initClusterNames() {
	for idx, val := range p.Steps {
		if val.ClusterName == "" {
			p.Steps[idx].ClusterName = p.ClusterName
		}
	}
}

type RequestSchemaValidator struct {
	argoAuthenticatorApi argoauthenticator.ArgoAuthenticatorApi
	repoManagerApi       reposerver.RepoManagerApi
}

func New(argoAuthenticatorApi argoauthenticator.ArgoAuthenticatorApi, repoApi reposerver.RepoManagerApi) RequestSchemaValidator {
	return RequestSchemaValidator{
		argoAuthenticatorApi: argoAuthenticatorApi,
		repoManagerApi:       repoApi,
	}
}

func (r RequestSchemaValidator) GetStepApplicationPath(orgName string, teamName string, gitRepoUrl string, gitCommitHash string, step string) string {
	pipelineData := r.getPipelineData(orgName, teamName, gitRepoUrl, gitCommitHash)
	for _, stepData := range pipelineData.Steps {
		if stepData.Name == step {
			return stepData.ApplicationPath
		}
	}
	return ""
}

func (r RequestSchemaValidator) GetStepApplicationPayload(orgName string, teamName string, gitRepoUrl string, gitCommitHash string, step string) (string, string) {
	var applicationPath string
	var clusterName string
	pipelineData := r.getPipelineData(orgName, teamName, gitRepoUrl, gitCommitHash)
	for _, stepData := range pipelineData.Steps {
		if stepData.Name == step {
			applicationPath = stepData.ApplicationPath
			clusterName = stepData.ClusterName
			break
		}
	}
	if applicationPath == "" {
		panic("This step does not have an application deployed")
	}
	request := reposerver.GetFileRequest{GitRepoUrl: gitRepoUrl, Filename: applicationPath, GitCommitHash: reposerver.RootCommit}
	return r.repoManagerApi.GetFileFromRepo(request, orgName, teamName), clusterName
}

func (r RequestSchemaValidator) ValidateSchemaAccess(orgName string, teamName string, gitRepoUrl string, gitCommitHash string, actionResourceEntries ...string) bool {
	pipelineData := r.getPipelineData(orgName, teamName, gitRepoUrl, gitCommitHash)
	for _, step := range pipelineData.Steps {
		if step.ClusterName == "" {
			panic(fmt.Sprintf("No cluster name specified for step %s", step.Name))
		}
		var i int
		for i = 0; i < len(actionResourceEntries); i += 2 {
			action := actionResourceEntries[i]
			resource := actionResourceEntries[i+1]
			if argoauthenticator.RbacResource(resource) == argoauthenticator.ClusterResource {
				if !r.argoAuthenticatorApi.CheckRbacPermissions(argoauthenticator.RbacAction(action), argoauthenticator.ClusterResource, step.ClusterName) {
					return false
				}
			} else if argoauthenticator.RbacResource(resource) == argoauthenticator.ApplicationResource {
				applicationSubresource := r.getArgoApplicationProjectAndName(orgName, teamName, gitRepoUrl, gitCommitHash, step.ApplicationPath)
				if !r.argoAuthenticatorApi.CheckRbacPermissions(argoauthenticator.RbacAction(action), argoauthenticator.ApplicationResource, applicationSubresource) {
					return false
				}
			}
		}
	}
	return true
}

func (r RequestSchemaValidator) VerifyRbac(action argoauthenticator.RbacAction, resource argoauthenticator.RbacResource, subresource string) bool {
	return r.argoAuthenticatorApi.CheckRbacPermissions(action, resource, subresource)
}

func (r RequestSchemaValidator) getPipelineData(orgName string, teamName string, gitRepoUrl string, gitCommitHash string) PipelineData {
	request := reposerver.GetFileRequest{GitRepoUrl: gitRepoUrl, Filename: reposerver.PipelineFileName, GitCommitHash: gitCommitHash}
	payload := r.repoManagerApi.GetFileFromRepo(request, orgName, teamName)
	var pipelineData PipelineData
	err := yaml.Unmarshal([]byte(payload), &pipelineData)
	if err != nil {
		panic(err)
	}
	pipelineData.initClusterNames()
	return pipelineData
}

func (r RequestSchemaValidator) getArgoApplicationProjectAndName(orgName string, teamName string, gitRepoUrl string, gitCommitHash string, applicationPath string) string {
	request := reposerver.GetFileRequest{GitRepoUrl: gitRepoUrl, Filename: applicationPath, GitCommitHash: gitCommitHash}
	argoApplicationConfig := r.repoManagerApi.GetFileFromRepo(request, orgName, teamName)
	var rawPayload map[interface{}]interface{}
	err := yaml.Unmarshal([]byte(argoApplicationConfig), &rawPayload)
	if err != nil {
		panic(err)
	}
	project := rawPayload["spec"].(map[interface{}]interface{})["project"]
	name := rawPayload["metadata"].(map[interface{}]interface{})["name"]
	if name == nil || name.(string) == "" {
		panic("the Argo CD app does not have a name")
	}
	if project == nil || project.(string) == "" {
		return defaultProject + "/" + name.(string)
	} else {
		return project.(string) + "/" + name.(string)
	}
}

func (r RequestSchemaValidator) GetArgoApplicationNamespace(argoApplicationConfig string) string {
	var rawPayload map[interface{}]interface{}
	err := yaml.Unmarshal([]byte(argoApplicationConfig), &rawPayload)
	if err != nil {
		panic(err)
	}
	namespace := rawPayload["spec"].(map[interface{}]interface{})["destination"].(map[interface{}]interface{})["namespace"]
	if namespace == nil || namespace.(string) == "" {
		panic("the Argo CD app does not have a namespace")
	}
	return namespace.(string)
}