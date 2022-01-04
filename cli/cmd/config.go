package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	confFileSuffix  string      = "./atlas/.atlas.yaml"
	certFileSuffix  string      = "./atlas/cert.pem"
	userPermissions os.FileMode = 0700
)

type conf struct {
	AtlasUrl   string `yaml:"URL"`
	Org        string `yaml:"ORG"`
	TLSEnabled bool   `yaml:"tls_enabled"`
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
		home, err := os.UserHomeDir()
		confFilename := home + confFileSuffix
		var configStruct conf

		cobra.CheckErr(err)
		if _, err := os.Stat(confFilename); os.IsNotExist(err) {
			_, err := os.Create(confFilename)
			if err != nil {
				fmt.Println("Unable to update atlas config.")
				return
			}
			configStruct.AtlasUrl = "localhost:8081"
			configStruct.Org = "org"
		}

		yamlFile, err := ioutil.ReadFile(confFilename)
		if err != nil {
			fmt.Println("Unable to update atlas config.")
			return
		}

		err = yaml.Unmarshal(yamlFile, &configStruct)
		if err != nil {
			fmt.Println("Unable to update atlas config: ", err)
			return
		}

		if cmd.Flags().Lookup("atlas.url").Changed {
			configStruct.AtlasUrl, _ = cmd.Flags().GetString("atlas.url")
		}

		if cmd.Flags().Lookup("atlas.org").Changed {
			configStruct.Org, _ = cmd.Flags().GetString("atlas.org")
		}

		// TODO: check that if we passing empty value it deletes cert file
		if cmd.Flags().Lookup("atlas.tls").Changed {
			certPEM, err := cmd.Flags().GetString("atlas.tls")
			if err != nil {
				fmt.Printf("failed to update atlas tls config: %s", err.Error())
			}
			if err := updateTLSConfig(certPEM, &configStruct); err != nil {
				fmt.Printf("failed to update atlas tls config: %s", err.Error())
				return
			}
		}

		data, _ := yaml.Marshal(&configStruct)
		err2 := ioutil.WriteFile(confFilename, data, 0)
		if err2 != nil {
			fmt.Println("Unable to update atlas config:", err2)
		}
	},
}

func updateTLSConfig(certPEM string, configStruct *conf) error {
	if certPEM != "" {
		configStruct.TLSEnabled = false
		return deleteCertFile()
	}
	if err := updateCertFile([]byte(certPEM)); err != nil {
		return err
	}
	configStruct.TLSEnabled = true
	return nil
}

func getCertFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home + "./atlas/cert.pem", nil
}

func checkCertFile() (bool, error) {
	certFilePath, err := getCertFilePath()
	if err != nil {
		return false, err
	}
	if _, err = os.Stat(certFilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func updateCertFile(data []byte) error {
	certFilePath, err := getCertFilePath()
	if err != nil {
		return err
	}
	if _, err = checkCertFile(); err != nil {
		return err
	}
	return ioutil.WriteFile(certFilePath, data, userPermissions)
}

func deleteCertFile() error {
	certFilePath, err := getCertFilePath()
	if err != nil {
		return err
	}
	if _, err = checkCertFile(); err != nil {
		return err
	}
	return os.Remove(certFilePath)
}

func getCertPEM() ([]byte, error) {
	path, err := getCertFilePath()
	if err != nil {
		return nil, err
	}
	exists, err := checkCertFile()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Println("failed to close config file after data reading: ", err)
		}
	}()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.PersistentFlags().StringP("atlas.url", "", "", "url of atlas")
	configCmd.PersistentFlags().StringP("atlas.org", "", "", "name of highest level org")
}
