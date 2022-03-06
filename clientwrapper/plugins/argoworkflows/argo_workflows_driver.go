package argoworkflows

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-cd/util/config"
	"github.com/argoproj/argo-workflows/v3"
	"github.com/argoproj/argo-workflows/v3/pkg/apiclient"
	"github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflow"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/util/kubeconfig"
	"greenops.io/client/progressionchecker/datamodel"
	corev1 "k8s.io/api/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"strings"
)

const (
	EnvArgoWorkflowServer     string = "ARGO_SERVER"
	DefaultWfApiServerAddress string = "argo-server.argo.svc.cluster.local:2746"
	EnvKubeconfigPath         string = "KUBECONFIG_PATH"
	ArgoWorkflowsNamespace    string = "argo"
)

type ArgoWfClientDriver struct {
	client apiclient.Client
}

func New() (*ArgoWfClientDriver, error) {
	apiServerAddress, kubeconfigPath := getWfClientCreationData()
	argoClient, err := getArgoWfClient(apiServerAddress, kubeconfigPath)
	if err != nil {
		log.Printf("Could not create Workflows client: %s", err)
		return nil, err
	}

	return &ArgoWfClientDriver{argoClient}, nil
}

func getArgoWfClient(apiServerAddress string, configPath string) (apiclient.Client, error) {
	ctx, client, err := apiclient.NewClientFromOpts(
		apiclient.Opts{
			//TODO: Additional ArgoServerOpts will have to be added
			ArgoServerOpts: apiclient.ArgoServerOpts{URL: apiServerAddress, Secure: true, InsecureSkipVerify: true},
			AuthSupplier: func() string {
				return GetAuthString(configPath)
			},
			ClientConfigSupplier: func() clientcmd.ClientConfig { return GetConfig(configPath) },
		})
	err = ctx.Err()
	if err != nil {
		log.Printf("Error while creating argo wf client %s", err)
		return nil, err
	}
	if err != nil {
		log.Printf("Error while creating argo wf client 0 %s", err)
	}
	return client, nil
}

func GetConfig(kubeconfigPath string) clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	loadingRules.ExplicitPath = kubeconfigPath
	overrides := clientcmd.ConfigOverrides{}
	return clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
}

func GetAuthString(kubeconfigPath string) string {
	token, ok := os.LookupEnv("ARGO_TOKEN")
	if ok {
		return token
	}
	restConfig, err := GetConfig(kubeconfigPath).ClientConfig()
	if err != nil {
		log.Printf("Error while creating argo wf client 1 %s", err)
		log.Fatal(err)
	}
	version := argo.GetVersion()
	restConfig = restclient.AddUserAgent(restConfig, fmt.Sprintf("argo-workflows/%s argo-cli", version.Version))
	authString, err := kubeconfig.GetAuthString(restConfig, kubeconfigPath)
	if err != nil {
		log.Printf("Error while creating argo wf client 2 %s", err)
		log.Fatal(err)
	}
	return authString
}

func (a *ArgoWfClientDriver) CreateAndDeploy(configPayload *string, variables []corev1.EnvVar) (string, string, error) {
	log.Printf("Deploying Argo Workflow...")
	workflowPayload := makeWorkflow(configPayload)
	for _, variable := range variables {
		if variable.ValueFrom == nil {
			workflowPayload.Spec.Arguments.Parameters = append(
				workflowPayload.Spec.Arguments.Parameters,
				wfv1.Parameter{Name: variable.Name, Value: wfv1.AnyStringPtr(variable.Value)},
			)
		}
	}
	client := a.client.NewWorkflowServiceClient()
	namespace := workflowPayload.Namespace
	if namespace == "" {
		namespace = ArgoWorkflowsNamespace
	}
	createdWorkflow, err := client.CreateWorkflow(context.TODO(), &workflow.WorkflowCreateRequest{
		Namespace:     namespace,
		Workflow:      &workflowPayload,
		ServerDryRun:  false,
		CreateOptions: nil,
	})
	if err != nil {
		log.Printf("Deploying the Argo Workflow threw an error: %s", err)
		return "", "", err
	}
	log.Printf("Deployed Argo Workflow %s in namespace %s", createdWorkflow.Name, createdWorkflow.Namespace)
	return createdWorkflow.Name, createdWorkflow.Namespace, nil
	//if errors.IsAlreadyExists(err) {
	//	return
	//}
}

func (a *ArgoWfClientDriver) Cleanup(watchKey datamodel.WatchKey) error {
	//TODO: Add any components necessary
	return nil
}

func (a *ArgoWfClientDriver) GetWorkflowStatus(name string, namespace string) (wfv1.WorkflowStatus, error) {
	client := a.client.NewWorkflowServiceClient()
	fetchedWorkflow, err := client.GetWorkflow(context.TODO(), &workflow.WorkflowGetRequest{
		Name:       name,
		Namespace:  namespace,
		GetOptions: nil,
	})
	if err != nil {
		return wfv1.WorkflowStatus{}, err
	}
	return fetchedWorkflow.Status, nil
}

func makeWorkflow(configPayload *string) wfv1.Workflow {
	var argoWorkflow wfv1.Workflow
	err := config.UnmarshalReader(strings.NewReader(*configPayload), &argoWorkflow)
	if err != nil {
		log.Printf("The unmarshalling step threw an error. Error was %s\n", err)
		return wfv1.Workflow{}
	}
	return argoWorkflow
}

func getWfClientCreationData() (string, string) {
	argoCdServer := os.Getenv(EnvArgoWorkflowServer)
	if argoCdServer == "" {
		argoCdServer = DefaultWfApiServerAddress
	}
	kubeconfigPath := "/dev/null" //os.Getenv(EnvKubeconfigPath)
	return argoCdServer, kubeconfigPath
}
