package schemavalidation

import (
	"fmt"
	"github.com/greenopsinc/util/git"

	"gopkg.in/yaml.v2"
	"greenops.io/workflowtrigger/api/argo"
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

type ClusterNamespaceGroup struct {
	ClusterName string `json:"clusterName"`
	Namespace   string `json:"namespace"`
}

type ClusterNamespaceGroups struct {
	Groups []ClusterNamespaceGroup `json:"groups"`
}

func (p *PipelineData) initClusterNames() {
	for idx, val := range p.Steps {
		if val.ClusterName == "" {
			p.Steps[idx].ClusterName = p.ClusterName
		}
	}
}

type RequestSchemaValidator struct {
	argoAuthenticatorApi argo.ArgoAuthenticatorApi
	repoManagerApi       reposerver.RepoManagerApi
}

func New(argoAuthenticatorApi argo.ArgoAuthenticatorApi, repoApi reposerver.RepoManagerApi) RequestSchemaValidator {
	return RequestSchemaValidator{
		argoAuthenticatorApi: argoAuthenticatorApi,
		repoManagerApi:       repoApi,
	}
}

func (r RequestSchemaValidator) GetStepApplicationPath(orgName string, teamName string, gitRepoSchemaInfo git.GitRepoSchemaInfo, gitCommitHash string, step string) string {
	pipelineData := r.getPipelineData(orgName, teamName, gitRepoSchemaInfo, gitCommitHash)
	for _, stepData := range pipelineData.Steps {
		if stepData.Name == step {
			return stepData.ApplicationPath
		}
	}
	return ""
}

func (r RequestSchemaValidator) hasStepApplicationPayload(pipelineData PipelineData, step string) bool {
	var applicationPath string
	for _, stepData := range pipelineData.Steps {
		if stepData.Name == step {
			applicationPath = stepData.ApplicationPath
			break
		}
	}
	if applicationPath == "" {
		return false
	}
	return true
}

func (r RequestSchemaValidator) GetStepApplicationPayloadWithoutPipelineData(orgName string, teamName string, gitRepoSchemaInfo git.GitRepoSchemaInfo, gitCommitHash string, step string) (string, string) {
	pipelineData := r.getPipelineData(orgName, teamName, gitRepoSchemaInfo, gitCommitHash)
	return r.GetStepApplicationPayload(pipelineData, orgName, teamName, gitRepoSchemaInfo, step)
}

func (r RequestSchemaValidator) GetStepApplicationPayload(pipelineData PipelineData, orgName string, teamName string, gitRepoSchemaInfo git.GitRepoSchemaInfo, step string) (string, string) {
	var applicationPath string
	var clusterName string
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
	request := git.GetFileRequest{GitRepoSchemaInfo: gitRepoSchemaInfo, Filename: applicationPath, GitCommitHash: reposerver.RootCommit}
	return r.repoManagerApi.GetFileFromRepo(request, orgName, teamName), clusterName
}

func (r RequestSchemaValidator) ValidateSchemaAccess(orgName string, teamName string, gitRepoSchemaInfo git.GitRepoSchemaInfo, gitCommitHash string, actionResourceEntries ...string) bool {
	pipelineData := r.getPipelineData(orgName, teamName, gitRepoSchemaInfo, gitCommitHash)
	for _, step := range pipelineData.Steps {
		if step.ClusterName == "" {
			panic(fmt.Sprintf("No cluster name specified for step %s", step.Name))
		}
		var i int
		for i = 0; i < len(actionResourceEntries); i += 2 {
			action := actionResourceEntries[i]
			resource := actionResourceEntries[i+1]
			if argo.RbacResource(resource) == argo.ClusterResource {
				if !r.VerifyRbac(argo.RbacAction(action), argo.ClusterResource, step.ClusterName) {
					return false
				}
			} else if argo.RbacResource(resource) == argo.ApplicationResource {
				if step.ApplicationPath != "" {
					applicationSubresource := r.getArgoApplicationProjectAndName(orgName, teamName, gitRepoSchemaInfo, gitCommitHash, step.ApplicationPath)
					if !r.VerifyRbac(argo.RbacAction(action), argo.ApplicationResource, applicationSubresource) {
						return false
					}
				}
			}
		}
	}
	return true
}

func (r RequestSchemaValidator) VerifyRbac(action argo.RbacAction, resource argo.RbacResource, subresource string) bool {
	return r.argoAuthenticatorApi.CheckRbacPermissions(action, resource, subresource)
}

func (r RequestSchemaValidator) getPipelineData(orgName string, teamName string, gitRepoSchemaInfo git.GitRepoSchemaInfo, gitCommitHash string) PipelineData {
	request := git.GetFileRequest{GitRepoSchemaInfo: gitRepoSchemaInfo, Filename: reposerver.PipelineFileName, GitCommitHash: gitCommitHash}
	payload := r.repoManagerApi.GetFileFromRepo(request, orgName, teamName)
	var pipelineData PipelineData
	err := yaml.Unmarshal([]byte(payload), &pipelineData)
	if err != nil {
		panic(err)
	}
	pipelineData.initClusterNames()
	return pipelineData
}

func (r RequestSchemaValidator) GetClusterNamespaceCombinations(orgName string, teamName string, gitRepoSchemaInfo git.GitRepoSchemaInfo, gitCommitHash string) ClusterNamespaceGroups {
	data := r.getPipelineData(orgName, teamName, gitRepoSchemaInfo, gitCommitHash)
	marked := make(map[string]bool)
	var groups ClusterNamespaceGroups

	for _, step := range data.Steps {
		if !r.hasStepApplicationPayload(data, step.Name) {
			continue
		}
		payload, _ := r.GetStepApplicationPayload(data, orgName, teamName, gitRepoSchemaInfo, step.Name)
		stepNamespace := r.GetArgoApplicationNamespace(payload)
		var namespace string
		if stepNamespace == "" {
			namespace = defaultProject
		} else {
			namespace = stepNamespace
		}
		if _, ok := marked[step.ClusterName+"_"+namespace]; ok {
			continue
		}
		groups.Groups = append(groups.Groups, ClusterNamespaceGroup{step.ClusterName, namespace})
		marked[step.ClusterName+"_"+namespace] = true
	}

	return groups
}

func (r RequestSchemaValidator) getArgoApplicationProjectAndName(orgName string, teamName string, gitRepoSchemaInfo git.GitRepoSchemaInfo, gitCommitHash string, applicationPath string) string {
	request := git.GetFileRequest{GitRepoSchemaInfo: gitRepoSchemaInfo, Filename: applicationPath, GitCommitHash: gitCommitHash}
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
