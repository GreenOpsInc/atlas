package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type GitCredOpen struct {
	Type string `json:"type"`
}
type GitCredToken struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}
type GitCredMachineUser struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
}
type GitRepoSchemaOpen struct {
	GitRepo    string      `json:"gitRepo"`
	PathToRoot string      `json:"pathToRoot"`
	GitCred    GitCredOpen `json:"gitCred"`
}

type GitRepoSchemaToken struct {
	GitRepo    string       `json:"gitRepo"`
	PathToRoot string       `json:"pathToRoot"`
	GitCred    GitCredToken `json:"gitCred"`
}

type GitRepoSchemaMachineUser struct {
	GitRepo    string             `json:"gitRepo"`
	PathToRoot string             `json:"pathToRoot"`
	GitCred    GitCredMachineUser `json:"gitCred"`
}

type UpdateTeamRequest struct {
	TeamName       string `json:"teamName"`
	ParentTeamName string `json:"parentTeamName"`
}

type ClusterSchema struct {
	ClusterIP   string `json:"clusterIP"`
	ExposedPort int    `json:"exposedPort"`
	ClusterName string `json:"clusterName"`
}

type NoDeployInfo struct {
	Name      string `json:"name"`
	Reason    string `json:"reason"`
	Namespace string `json:"namespace"`
}

var cfgFile string
var atlasURL string
var orgName string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "atlas",
	Short: "A brief description of your application",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.atlas.yaml)")

	rootCmd.PersistentFlags().StringVar(&atlasURL, "url", "localhost:8081", "override the default atlas url for a single command execution ")
	rootCmd.PersistentFlags().StringVar(&orgName, "org", "org", "override the default org name for a single command execution")
}

func bindFlags() {
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		viper.BindEnv(f.Name, strings.ToUpper(f.Name))
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			rootCmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func getHttpClient() *http.Client {
	certPEM, err := getCertPEM()
	if err != nil {
		log.Fatal(err)
	}

	clientTLSConf := &tls.Config{}
	if certPEM == nil {
		clientTLSConf.InsecureSkipVerify = true
	} else {
		certpool := bestEffortSystemCertPool()
		certpool.AppendCertsFromPEM(certPEM)
		clientTLSConf.RootCAs = certpool
	}

	return &http.Client{Timeout: 20 * time.Second, Transport: &http.Transport{
		TLSClientConfig: clientTLSConf,
	}}
}

func bestEffortSystemCertPool() *x509.CertPool {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		log.Println("root ca not found, returning new...")
		return x509.NewCertPool()
	}
	log.Println("root ca found")
	return rootCAs
}
