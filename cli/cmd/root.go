package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strings"
)
type GitCredOpen struct {
	Type string `json:"type"`
}
type GitCredToken struct {
	Type string `json:"type"`
	Token string `json:"token"`
}
type GitCredMachineUser struct {
	Type string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
}
type GitRepoSchemaOpen struct {
	GitRepo string `json:"gitRepo"`
	PathToRoot string `json:"pathToRoot"`
	GitCred GitCredOpen `json:"gitCred"`
}

type GitRepoSchemaToken struct {
	GitRepo string `json:"gitRepo"`
	PathToRoot string `json:"pathToRoot"`
	GitCred GitCredToken `json:"gitCred"`
}

type GitRepoSchemaMachineUser struct {
	GitRepo string `json:"gitRepo"`
	PathToRoot string `json:"pathToRoot"`
	GitCred GitCredMachineUser `json:"gitCred"`
}

type UpdateTeamRequest struct {
	TeamName string `json:"teamName"`	
	ParentTeamName string `json:"parentTeamName"`
}

type ClusterSchema struct {
	ClusterIP string `json:"clusterIP"`
	ExposedPort int `json:"exposedPort"`
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
	Long: ``,
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
	rootCmd.PersistentFlags().StringVar(&orgName, "org","org", "override the default org name for a single command execution")
	
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

func bindFlags(){
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		viper.BindEnv(f.Name, strings.ToUpper(f.Name))
		if !f.Changed && viper.IsSet(f.Name){
			val := viper.Get(f.Name)
			rootCmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}


