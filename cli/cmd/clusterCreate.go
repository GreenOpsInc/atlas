package cmd

import (
	"bytes"
	"encoding/json"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/common"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/util/clusterauth"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/greenopsinc/util/cluster"
	"io"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	// "strconv"
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"time"
)

// NewClusterAddCommand returns a new instance of an `argocd cluster add` command
func NewClusterAddCommand(pathOpts *clientcmd.PathOptions) *cobra.Command {
	var (
		serviceAccount          string
		awsRoleArn              string
		awsClusterName          string
		systemNamespace         string
		namespaces              []string
		name                    string
		execProviderCommand     string
		execProviderArgs        []string
		execProviderEnv         map[string]string
		execProviderAPIVersion  string
		execProviderInstallHint string
		server                  string
	)
	var command = &cobra.Command{
		Use:   "create CONTEXT",
		Short: "atlas cluster create CONTEXT",
		Run: func(c *cobra.Command, args []string) {
			var clst *v1alpha1.Cluster
			if server == "" {
				//ArgoCD logic for cluster config fetching
				var configAccess clientcmd.ConfigAccess = pathOpts
				if len(args) == 0 {
					log.Print("Choose a context name from:")
					printKubeContexts(configAccess)
					os.Exit(1)
				}
				config, err := configAccess.GetStartingConfig()
				errors.CheckError(err)
				contextName := args[0]
				clstContext := config.Contexts[contextName]
				if clstContext == nil {
					log.Fatalf("Context %s does not exist in kubeconfig", contextName)
				}

				overrides := clientcmd.ConfigOverrides{
					Context: *clstContext,
				}
				clientConfig := clientcmd.NewDefaultClientConfig(*config, &overrides)
				conf, err := clientConfig.ClientConfig()
				errors.CheckError(err)

				managerBearerToken := ""
				var awsAuthConf *v1alpha1.AWSAuthConfig
				var execProviderConf *v1alpha1.ExecProviderConfig
				if awsClusterName != "" {
					awsAuthConf = &v1alpha1.AWSAuthConfig{
						ClusterName: awsClusterName,
						RoleARN:     awsRoleArn,
					}
				} else if execProviderCommand != "" {
					execProviderConf = &v1alpha1.ExecProviderConfig{
						Command:     execProviderCommand,
						Args:        execProviderArgs,
						Env:         execProviderEnv,
						APIVersion:  execProviderAPIVersion,
						InstallHint: execProviderInstallHint,
					}
				} else {
					// Install RBAC resources for managing the cluster
					clientset, err := kubernetes.NewForConfig(conf)
					errors.CheckError(err)
					if serviceAccount != "" {
						managerBearerToken, err = clusterauth.GetServiceAccountBearerToken(clientset, systemNamespace, serviceAccount)
					} else {
						managerBearerToken, err = clusterauth.InstallClusterManagerRBAC(clientset, systemNamespace, namespaces)
					}
					errors.CheckError(err)
				}
				if name != "" {
					contextName = name
				}
				clst = newCluster(contextName, namespaces, conf, managerBearerToken, awsAuthConf, execProviderConf)
			}
			//Atlas logic for token fetching/setting & request making
			var body cluster.ClusterCreateRequest
			if server != "" {
				body = cluster.ClusterCreateRequest{Server: server, Name: name, Config: nil}
			} else {
				body = cluster.ClusterCreateRequest{Server: clst.Server, Name: clst.Name, Config: &clst.Config}
				name = clst.Name
			}
			defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
			errors.CheckError(err)
			localConfig, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
			context, _ := localConfig.ResolveContext(apiclient.ClientOptions{}.Context)
			url := "http://" + atlasURL + "/cluster/" + orgName
			var req *http.Request
			json, _ := json.Marshal(body)
			req, _ = http.NewRequest("POST", url, bytes.NewBuffer(json))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))

			client := &http.Client{Timeout: 20 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Request failed with the following error:", err)
				return
			}
			statusCode := resp.StatusCode
			errBody, _ := io.ReadAll(resp.Body)
			if statusCode == 200 {
				fmt.Printf("Successfully created cluster %s for org %s", name, orgName)
			} else {
				fmt.Printf("Error creating cluster: %s", errBody)
			}
		},
	}
	command.Flags().StringVar(&serviceAccount, "service-account", "", fmt.Sprintf("System namespace service account to use for kubernetes resource management. If not set then default \"%s\" SA will be created", clusterauth.ArgoCDManagerServiceAccount))
	command.Flags().StringVar(&awsClusterName, "aws-cluster-name", "", "AWS Cluster name if set then aws cli eks token command will be used to access cluster")
	command.Flags().StringVar(&awsRoleArn, "aws-role-arn", "", "Optional AWS role arn. If set then AWS IAM Authenticator assume a role to perform cluster operations instead of the default AWS credential provider chain.")
	command.Flags().StringVar(&systemNamespace, "system-namespace", common.DefaultSystemNamespace, "Use different system namespace")
	command.Flags().StringVar(&name, "name", "", "Overwrite the cluster name")
	command.Flags().StringVar(&execProviderCommand, "exec-command", "", "Command to run to provide client credentials to the cluster. You may need to build a custom ArgoCD image to ensure the command is available at runtime.")
	command.Flags().StringArrayVar(&execProviderArgs, "exec-command-args", nil, "Arguments to supply to the --exec-command command")
	command.Flags().StringToStringVar(&execProviderEnv, "exec-command-env", nil, "Environment vars to set when running the --exec-command command")
	command.Flags().StringVar(&execProviderAPIVersion, "exec-command-api-version", "", "Preferred input version of the ExecInfo for the --exec-command")
	command.Flags().StringVar(&execProviderInstallHint, "exec-command-install-hint", "", "Text shown to the user when the --exec-command executable doesn't seem to be present")
	command.Flags().StringVar(&server, "server", "", "Server IP")
	return command
}

func printKubeContexts(ca clientcmd.ConfigAccess) {
	config, err := ca.GetStartingConfig()
	errors.CheckError(err)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() { _ = w.Flush() }()
	columnNames := []string{"CURRENT", "NAME", "CLUSTER", "SERVER"}
	_, err = fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))
	errors.CheckError(err)

	// sort names so output is deterministic
	contextNames := make([]string, 0)
	for name := range config.Contexts {
		contextNames = append(contextNames, name)
	}
	sort.Strings(contextNames)

	if config.Clusters == nil {
		return
	}

	for _, name := range contextNames {
		// ignore malformed kube config entries
		context := config.Contexts[name]
		if context == nil {
			continue
		}
		cluster := config.Clusters[context.Cluster]
		if cluster == nil {
			continue
		}
		prefix := " "
		if config.CurrentContext == name {
			prefix = "*"
		}
		_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", prefix, name, context.Cluster, cluster.Server)
		errors.CheckError(err)
	}
}

func newCluster(name string, namespaces []string, conf *rest.Config, managerBearerToken string, awsAuthConf *v1alpha1.AWSAuthConfig, execProviderConf *v1alpha1.ExecProviderConfig) *v1alpha1.Cluster {
	tlsClientConfig := v1alpha1.TLSClientConfig{
		Insecure:   conf.TLSClientConfig.Insecure,
		ServerName: conf.TLSClientConfig.ServerName,
		CAData:     conf.TLSClientConfig.CAData,
		CertData:   conf.TLSClientConfig.CertData,
		KeyData:    conf.TLSClientConfig.KeyData,
	}
	if len(conf.TLSClientConfig.CAData) == 0 && conf.TLSClientConfig.CAFile != "" {
		data, err := ioutil.ReadFile(conf.TLSClientConfig.CAFile)
		errors.CheckError(err)
		tlsClientConfig.CAData = data
	}
	if len(conf.TLSClientConfig.CertData) == 0 && conf.TLSClientConfig.CertFile != "" {
		data, err := ioutil.ReadFile(conf.TLSClientConfig.CertFile)
		errors.CheckError(err)
		tlsClientConfig.CertData = data
	}
	if len(conf.TLSClientConfig.KeyData) == 0 && conf.TLSClientConfig.KeyFile != "" {
		data, err := ioutil.ReadFile(conf.TLSClientConfig.KeyFile)
		errors.CheckError(err)
		tlsClientConfig.KeyData = data
	}

	clst := v1alpha1.Cluster{
		Server:     conf.Host,
		Name:       name,
		Namespaces: namespaces,
		Config: v1alpha1.ClusterConfig{
			TLSClientConfig:    tlsClientConfig,
			AWSAuthConfig:      awsAuthConf,
			ExecProviderConfig: execProviderConf,
		},
	}

	// Bearer token will preferentially be used for auth if present,
	// Even in presence of key/cert credentials
	// So set bearer token only if the key/cert data is absent
	if len(tlsClientConfig.CertData) == 0 || len(tlsClientConfig.KeyData) == 0 {
		clst.Config.BearerToken = managerBearerToken
	}

	return &clst
}

func init() {
	pathOpts := clientcmd.NewDefaultPathOptions()
	clusterCreateCmd := NewClusterAddCommand(pathOpts)
	clusterCmd.AddCommand(clusterCreateCmd)
	clusterCreateCmd.PersistentFlags().StringVar(&pathOpts.LoadingRules.ExplicitPath, pathOpts.ExplicitFileFlag, pathOpts.LoadingRules.ExplicitPath, "use a particular kubeconfig file")
}
