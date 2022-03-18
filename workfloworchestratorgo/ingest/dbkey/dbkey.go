package dbkey

import "strings"

func MakeDbTeamKey(orgName, teamName string) string {
	return orgName + "-" + teamName
}

func MakeDbStepKey(orgName, teamName, pipelineName, stepName string) string {
	return strings.Join([]string{orgName, teamName, pipelineName, stepName}, "-")
}

func MakeClientRequestQueueKey(orgName, clusterName string) string {
	return strings.Join([]string{orgName, clusterName, "events"}, "-")
}

func MakeDbMetadataKey(orgName, teamName, pipelineName, stepName string) string {
	return strings.Join([]string{orgName, teamName, pipelineName, stepName, "meta"}, "-")
}

func MakeDbPipelineInfoKey(orgName, teamName, pipelineName string) string {
	return strings.Join([]string{orgName, teamName, pipelineName, "meta"}, "-")
}

func MakeDbListOfTeamsKey(orgName string) string {
	return orgName + "-teams"
}

func MakeDbClusterKey(orgName, clusterName string) string {
	return orgName + "-" + clusterName
}
