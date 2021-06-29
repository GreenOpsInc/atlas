package progressionchecker

import (
	"bytes"
	"encoding/json"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"greenops.io/client/k8sdriver"
	"greenops.io/client/progressionchecker/datamodel"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	EndTransactionMarker         string = "__ATLAS_END_TRANSACTION_MARKER__"
	ProgressionChannelBufferSize int    = 100
	DefaultMetricsServerAddress  string = "http://argocd-metrics.argocd.svc.cluster.local:8082/metrics"
	//Command to get localhost address: "minikube ssh 'grep host.minikube.internal /etc/hosts | cut -f1'"
	DefaultWorkflowTriggerAddress string = "http://atlasworkflowtrigger.default.svc.cluster.local"
	MetricsServerEnvVar           string = "ARGOCD_METRICS_SERVER_ADDR"
	WorkflowTriggerEnvVar         string = "WORKFLOW_TRIGGER_SERVER_ADDR"
	HttpRequestRetryLimit         int    = 3
)

func listenForApplicationsToWatch(inputChannel chan string, outputChannel chan string) {
	var listOfWatchKeys []string
	for {
		for {
			newWatchKey := <-inputChannel
			if newWatchKey == EndTransactionMarker {
				break
			}
			listOfWatchKeys = append(listOfWatchKeys, newWatchKey)
		}
		for len(listOfWatchKeys) > 0 {
			outputChannel <- listOfWatchKeys[0]
			listOfWatchKeys = listOfWatchKeys[1:]
		}
	}
}

func checkForCompletedApplications(kubernetesClient k8sdriver.KubernetesClientGetRestricted, channel chan string, metricsServerAddress string, workflowTriggerAddress string) {
	watchedApplications := make(map[datamodel.WatchKey]string)
	for {
		//Update watched applications list with new entries
		for {
			doneReading := false
			select {
			case newKey := <-channel:
				var newWatchKey datamodel.WatchKey
				_ = json.NewDecoder(strings.NewReader(newKey)).Decode(&newWatchKey)
				if _, ok := watchedApplications[newWatchKey]; !ok {
					watchedApplications[newWatchKey] = datamodel.Missing
					log.Printf("Added key %s to watched list\n", newWatchKey)
				}
			default:
				//No more items to read in from channel
				doneReading = true
			}
			if doneReading {
				break
			}
		}
		//Check health of watched applications
		var argoAppInfo []datamodel.ArgoAppMetricInfo
		for i := 0; i < HttpRequestRetryLimit; i++ {
			argoAppInfo = getUnmarshalledMetrics(metricsServerAddress)
			if argoAppInfo != nil {
				break
			}
		}

		for _, appInfo := range argoAppInfo {
			key := datamodel.WatchKey{}
			for mapKey, _ := range watchedApplications {
				if mapKey.Name == appInfo.Name && mapKey.Namespace == appInfo.Namespace {
					key = mapKey
					break
				}
			}
			if key.PipelineName == "" {
				log.Printf("The Argo application named %s was never marked as \"watched\"", appInfo.Name)
				continue
			}
			if appInfo.HealthStatus == string(health.HealthStatusHealthy) && appInfo.SyncStatus == string(v1alpha1.SyncStatusCodeSynced) {
				watchedApplications[key] = datamodel.Healthy
				eventInfo := datamodel.MakeApplicationEvent(key, appInfo)
				for i := 0; i < HttpRequestRetryLimit; i++ {
					if generateEvent(eventInfo, workflowTriggerAddress) {
						//TODO: This deletion is a TEMPORARY measure. It should not be done longer term, as we want to continue receiving Argo datamodel about application
						delete(watchedApplications, key)
						break
					}
				}
			} else {
				watchedApplications[key] = datamodel.NotHealthy
			}
		}

		for key, _ := range watchedApplications {
			log.Printf("Checking for key %s", key.Name)
			if key.Type == datamodel.WatchTestKey {
				jobStatus, numPods := kubernetesClient.GetJob(key.Name, key.Namespace)
				log.Printf("Job status succeeded: %d, status failed %d", jobStatus.Succeeded, jobStatus.Failed)
				if numPods == -1 {
					//TODO: GetJob failed. Should have better handling than continue
					continue
				}
				if jobStatus.Succeeded == numPods || jobStatus.Failed == numPods {
					eventInfo := datamodel.MakeTestEvent(key, jobStatus.Succeeded > 0)
					for i := 0; i < HttpRequestRetryLimit; i++ {
						if generateEvent(eventInfo, workflowTriggerAddress) {
							delete(watchedApplications, key)
							break
						}
					}
				}
			}
		}

		duration := 10 * time.Second
		time.Sleep(duration)
	}
}

func unmarshall(payload *string) []datamodel.ArgoAppMetricInfo {
	metricTypePattern, _ := regexp.Compile("(argocd_app_info{.*} [0-9]*)+")
	relevantPayload := metricTypePattern.FindAllString(*payload, -1)
	var argoAppInfo []datamodel.ArgoAppMetricInfo
	for _, payloadLine := range relevantPayload {
		metricVariablePattern, _ := regexp.Compile("([a-zA-Z_-]*)=\"([a-zA-Z:/_.-]*)\"")
		variableMap := make(map[string]string)
		for _, variable := range metricVariablePattern.FindAllStringSubmatch(payloadLine, -1) {
			variableMap[variable[1]] = variable[2]
		}
		//We should be using reflection in the future to assign all these values. For now this will work.
		appInfo := datamodel.ArgoAppMetricInfo{
			DestNamespace: variableMap["dest_namespace"],
			DestServer:    variableMap["dest_server"],
			HealthStatus:  variableMap["health_status"],
			Name:          variableMap["name"],
			Namespace:     variableMap["namespace"],
			Operation:     variableMap["operation"],
			Project:       variableMap["project"],
			Repo:          variableMap["repo"],
			SyncStatus:    variableMap["sync_status"],
		}
		argoAppInfo = append(argoAppInfo, appInfo)
	}
	return argoAppInfo
}

func getUnmarshalledMetrics(metricsServerAddress string) []datamodel.ArgoAppMetricInfo {
	resp, err := http.Get(metricsServerAddress)
	if err != nil {
		log.Printf("Error querying metrics server. Error was %s\n", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading in response body from metrics server. Error was %s\n", err)
		return nil
	}
	stringBody := string(body)
	return unmarshall(&stringBody)
}

func generateEvent(eventInfo datamodel.EventInfo, workflowTriggerAddress string) bool {
	data, err := json.Marshal(eventInfo)
	if err != nil {
		return false
	}
	resp, err := http.Post(workflowTriggerAddress+"/client/generateEvent", "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error generating progression event. Error was %s\n", err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode/100 == 2
}

func Start(kubernetesClient k8sdriver.KubernetesClientGetRestricted, channel chan string) {
	metricsServerAddress := os.Getenv(MetricsServerEnvVar)
	if metricsServerAddress == "" {
		metricsServerAddress = DefaultMetricsServerAddress
	}
	workflowTriggerAddress := os.Getenv(WorkflowTriggerEnvVar)
	if workflowTriggerAddress == "" {
		workflowTriggerAddress = DefaultWorkflowTriggerAddress
	}
	progressionCheckerChannel := make(chan string, ProgressionChannelBufferSize)
	//TODO: Add initial cache creation
	go listenForApplicationsToWatch(channel, progressionCheckerChannel)
	checkForCompletedApplications(kubernetesClient, progressionCheckerChannel, metricsServerAddress, workflowTriggerAddress)
}
