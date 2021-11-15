package plugins

import (
	"errors"
	"greenops.io/client/plugins/argoworkflows"
	"greenops.io/client/progressionchecker/datamodel"
	"log"
)

type PluginType string

const (
	ArgoWorkflow PluginType = "ArgoWorkflowTask"
)

type PluginInt interface {
	CreateAndDeploy(configPayload *string, variables map[string]string) (string, string, error)
	CheckStatus(watchKey datamodel.WatchKey) datamodel.EventInfo
}

type Plugin struct {
	Type         PluginType
	PluginObject PluginInt
}

type Plugins []Plugin

func New(pluginType PluginType) (Plugin, error) {
	if pluginType == ArgoWorkflow {
		var pluginObj PluginInt
		var err error
		pluginObj, err = argoworkflows.New()
		if err != nil {
			return Plugin{}, err
		}
		return Plugin{pluginType, pluginObj}, nil
	} else {
		return Plugin{}, errors.New("PluginType did not match existing list")
	}
}

func (pluginsList Plugins) GetPlugin(pluginType PluginType) (Plugin, error) {
	//For plugins, the PluginType and WatchKeyType should match each other
	for _, plugin := range pluginsList {
		if plugin.Type == pluginType {
			return plugin, nil
		}
	}
	plugin, err := New(pluginType)
	if err != nil {
		log.Printf("Could not get plugin: %s", err)
		return Plugin{}, err
	}
	return plugin, nil
}
