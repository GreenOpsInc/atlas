package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
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

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()

		cobra.CheckErr(err)

		// Search config in home directory with name ".atlas" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".atlas")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		bindFlags()
	}

	atlasURL, _ = rootCmd.Flags().GetString("url")
	orgName, _ = rootCmd.Flags().GetString("org")
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
	// TODO: find a way to somehow retrieve this certificate from kubernetes secrets
	certPEM := []byte(`-----BEGIN CERTIFICATE-----
MIIFPDCCAySgAwIBAgIRAP8psTlD2mFYWPTTCPL4WFgwDQYJKoZIhvcNAQELBQAw
JjELMAkGA1UEBhMCVVMxFzAVBgNVBAoTDkdyZWVuT3BzLCBJTkMuMB4XDTIxMTIz
MTA5MzE0MloXDTMxMTIzMTA5MzE0MlowJjELMAkGA1UEBhMCVVMxFzAVBgNVBAoT
DkdyZWVuT3BzLCBJTkMuMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA
7hS0+BrCuKYXTnCLTtRUDhEOKxB0Dohryp0ROeUr2N3ozvX6I2Yhq/wav9Ul6yt3
mrC397XACy1Ze9eE9zYeapB2VQgApXnQeVodv+1XP9vKz7eEeSE7ctymBQ819vQ9
MndOO+l/U1Fz9tYVSShAE6sRyepxfllc930m4vESyiY4IrXV7PsH63wmuSnXzEjG
kHMai7IZIyRa/iahpKo9sv9dcjURLeDrDiFWfv/lp+YcUaIqN0g+yWOKJs3zb4ba
U2yoW/SnBWrMJIldp3GEsRUIads/oC36kZB8gJ6IcoUthjkIZlNYeisjaVk1/GYJ
D1pS6CjbNiEZTxH+wEwzmiF1VKvM4+EEqdWIX4Cnh4NTJ41KX60JfonOmGJD3J0V
6eW6U6xC0gNA43hsjmiaIgpIjk2U1lkBex3D74rXKVrjC/yKAWVd/r3x6/I0ttdN
EfAChSlB/IXEXT3+3J2oy0zKFQEkAWwneiHYRZJfSfMMfoG7mb84UqAr6EXbNrgF
YyP7VB8nFSxQySYE7RH4Tnqhoiwuy6Yl1zDM9bXjL1SOiHfNKiBvla67ZeEvpZVw
XatBTpi9/qUXXaxSPzmLeu5KrcyxtW5oU6FS4Q7NnvPoY7bWnjwzxO03cxqygaI7
0HExIZvsR5nVjI02Mr9Gu4p8fdoW8FrCCqpBsONHuoECAwEAAaNlMGMwDgYDVR0P
AQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA4GA1UdDgQHBAUBAgMEBjAs
BgNVHREEJTAjgglsb2NhbGhvc3SHBH8AAAGHEAAAAAAAAAAAAAAAAAAAAAEwDQYJ
KoZIhvcNAQELBQADggIBAOxhBLmSUv1HgMgWsU0eMLzYiDKyvLMbFMM2zihnDYlq
5Cuul2jgErVEPfu72fjr1II83o0LcvTi0n9BbE9JhVP7mktuAHGEg8qmt248+XmO
8yqOuquxyqqfhlBEaXcq9cvnQ4Ac8hHXk2mpUfgPo+lm2w25gEJmhRAhxomflLrK
BtkCnRY3yp36WZ5mkcULRaLM7S0ZVOM+vjINcN9mqtcieRlCSRG6XcByRpb0XHnA
w6pY6MtAojmFOpNe7s2cfH9rIROMGaIc1uLGjX8hheBUuP6OBqxK/SBgBQX4HPas
JyUGshTJBwI5uLBaRIBF4ZQXGwnW8GmFDHnozDMDr8K2rjZcFsR9RpWQ0JJZzMGb
6lEjNdQJDQh2qeHl0LGdvzlio1r9Vyn+5PjCPy6Ky1xRjMnfwuMLahDvTpuIPfBz
783dne2UwYl64n1pS/2RDLWV0Y82dkyGk0613kFuTJwX+fqKEVyCDGk+e0mdxGj7
jyXohZdiiOoLN17HRe/7Sm3wbg7R+xKIcA2BBK6YWxR0dGwmMVixwXR5RIy0stHf
RsyluS5+YtPXGW+8gKriMYEEu9g3iRLj4/bPSk/u8sTpUN4F1WzShfY2UNYSWfxm
Lw9sMhgn3OyJQW27joqNxfhZort71qviiGr6ZuE5UXkPbcRpoDKLZ/PDKN+5GTa4
-----END CERTIFICATE-----`)

	certpool := bestEffortSystemCertPool()
	certpool.AppendCertsFromPEM(certPEM)
	clientTLSConf := &tls.Config{
		RootCAs: certpool,
	}
	transport := &http.Transport{
		TLSClientConfig: clientTLSConf,
	}

	return &http.Client{Timeout: 20 * time.Second, Transport: transport}
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
