package db

import (
	"encoding/hex"
	"strings"
)

func MakeDbTeamKey(orgName string, teamName string) string {
	return orgName + "-" + teamName
}

func MakeDbStepKey(orgName string, teamName string, pipelineName string, stepName string) string {
	return strings.Join([]string{orgName, teamName, pipelineName, stepName}, "-")
}

func MakeDbPipelineInfoKey(orgName string, teamName string, pipelineName string) string {
	return strings.Join([]string{orgName, teamName, pipelineName, "meta"}, "-")
}

func MakeDbListOfTeamsKey(orgName string) string {
	return orgName + "-teams"
}

func MakeDbListOfStepsKey(orgName string, teamName string, pipelineName string) string {
	return orgName + "-" + teamName + "-" + pipelineName + "-step"
}

func MakeDbClusterKey(orgName string, clusterName string) string {
	return orgName + "-" + clusterName
}

func MakeSecretName(orgName string, teamName string, pipelineName string) string {
	name := orgName + "-" + teamName + "-" + pipelineName + "-gitcred"
	return strings.ToLower(hex.EncodeToString([]byte(name)))
}