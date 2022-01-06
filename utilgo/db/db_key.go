package db

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func MakeDbNotificationKey(requestId string) string {
	return fmt.Sprintf("notification-%s", requestId)
}

func MakeDbTeamKey(orgName string, teamName string) string {
	return orgName + "-" + teamName
}

func MakeDbStepKey(orgName string, teamName string, pipelineName string, stepName string) string {
	return strings.Join([]string{orgName, teamName, pipelineName, stepName}, "-")
}

func MakeClientRequestQueueKey(orgName string, clusterName string) string {
	return strings.Join([]string{orgName, clusterName, "events"}, "-")
}

func MakeClientNotificationQueueKey(orgName string, clusterName string) string {
	return strings.Join([]string{orgName, clusterName, "notification"}, "-")
}

func MakeDbPipelineInfoKey(orgName string, teamName string, pipelineName string) string {
	return strings.Join([]string{orgName, teamName, pipelineName, "meta"}, "-")
}

func MakeDbListOfTeamsKey(orgName string) string {
	return orgName + "-teams"
}

func MakeDbClusterKey(orgName string, clusterName string) string {
	return orgName + "-" + clusterName
}

func MakeSecretName(orgName string, teamName string, pipelineName string) string {
	name := orgName + "-" + teamName + "-" + pipelineName + "-gitcred"
	return strings.ToLower(hex.EncodeToString([]byte(name)))
}
