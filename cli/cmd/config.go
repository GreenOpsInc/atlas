package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	
	
)

type conf struct {
	AtlasUrl string `yaml:"URL"`
	Org string `yaml:"ORG"`
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config <config value>",
	Short: "Command to set atlas configurations",
	Long: `
Command to set atlas configurations on a global level. Use the appropriate flag to
to set the respective configuration value. The default value for these config variables
will be updated for future commands after executing this command.`,
	Run: func(cmd *cobra.Command, args []string) {
		home,err:= os.UserHomeDir()
		filename:=home+"/.atlas.yaml"
		var configStruct conf
		
		cobra.CheckErr(err)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			_, err := os.Create(filename)
			if err!=nil{
				fmt.Println("Unable to update atlas config.")
				return
			}
			configStruct.AtlasUrl="localhost:8081"
			configStruct.Org="org"
		}

		yamlFile, err:= ioutil.ReadFile(filename)
		if err!=nil{
			fmt.Println("Unable to update atlas config.")
			return
		}

		err = yaml.Unmarshal(yamlFile, &configStruct)
		if err!=nil{
			fmt.Println("Unable to update atlas config: ",err)
			return
		}

		if cmd.Flags().Lookup("atlas.url").Changed{
			configStruct.AtlasUrl,_=cmd.Flags().GetString("atlas.url")
		}

		if cmd.Flags().Lookup("atlas.org").Changed{
			configStruct.Org,_=cmd.Flags().GetString("atlas.org")
		}

		data,_:=yaml.Marshal(&configStruct)
		err2:= ioutil.WriteFile(filename, data, 0)
		if err2!=nil{
			fmt.Println("Unable to update atlas config:", err2)	
		}

	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.PersistentFlags().StringP("atlas.url", "", "", "url of atlas")
	configCmd.PersistentFlags().StringP("atlas.org", "", "", "name of highest level org")
}
