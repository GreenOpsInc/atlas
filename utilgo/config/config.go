package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	userPermissions os.FileMode = 0777

	configPathEnvName string = "ATLAS_CONFIG_PATH"
	configDefaultPath string = "/home/.atlas/"
)

func GetConfigPath() (string, error) {
	path := configDefaultPath
	if val := os.Getenv(configPathEnvName); val != "" {
		path = val
	}

	if err := checkAndCreateConfigDirectory(path); err != nil {
		return "", err
	}
	return path, nil
}

func WriteDataToConfigFile(data []byte, path string) error {
	if err := checkConfigFile(path); err != nil {
		return err
	}

	dirPath := path[0:strings.LastIndex(path, "/")]
	if err := os.MkdirAll(dirPath, userPermissions); err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, userPermissions)
}

func ReadDataFromConfigFile(path string) ([]byte, error) {
	if err := checkConfigFile(path); err != nil {
		return nil, err
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

func DeleteConfigFile(path string) error {
	if err := checkConfigFile(path); err != nil {
		return err
	}
	return os.Remove(path)
}

func checkConfigFile(path string) error {
	confPath, err := GetConfigPath()
	if err != nil {
		return err
	}
	if !strings.Contains(path, confPath) {
		return errors.New("file is not located in the atlas config directory")
	}
	return nil
}

func checkAndCreateConfigDirectory(path string) error {
	if _, err := os.Stat(path); err == nil {
		log.Println("atlas configuration directory exists")
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return errors.New(fmt.Sprintf("failed to check atlas config directory: %s", err.Error()))
	}

	log.Println("atlas configuration directory is not exist, creating...")
	if err := os.MkdirAll(path, userPermissions); err != nil {
		return err
	}
	return nil
}
