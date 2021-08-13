package progressionchecker

import (
	"bytes"
	"encoding/json"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"greenops.io/client/argodriver"
	"greenops.io/client/k8sdriver"
	"greenops.io/client/progressionchecker/datamodel"
	"io"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	DefaultWorkflowTriggerAddress string = "http://workflowtrigger.default.svc.cluster.local:8080"
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

func checkForCompletedApplications(kubernetesClient k8sdriver.KubernetesClientGetRestricted, argoClient argodriver.ArgoGetRestrictedClient, channel chan string, metricsServerAddress string, workflowTriggerAddress string) {
	watchedApplications := make(map[string]datamodel.WatchKey)
	for {
		//Update watched applications list with new entries
		for {
			doneReading := false
			select {
			case newKey := <-channel:
				var newWatchKey datamodel.WatchKey
				_ = json.NewDecoder(strings.NewReader(newKey)).Decode(&newWatchKey)
				keyForWatchKey := newWatchKey.GetKeyFromWatchKey()
				watchedApplications[keyForWatchKey] = newWatchKey
				log.Printf("Added key %s to watched list\n", newWatchKey)
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

		deleteKeys := make([]string, 0)
		for mapKey, watchKey := range watchedApplications {
			if watchKey.Type == datamodel.WatchArgoApplicationKey {
				for _, appInfo := range argoAppInfo {
					if watchKey.Name == appInfo.Name && watchKey.Namespace == appInfo.Namespace {
						var eventInfo datamodel.EventInfo
						if appInfo.HealthStatus == health.HealthStatusProgressing {
							break
						} else if appInfo.SyncStatus == v1alpha1.SyncStatusCodeSynced {
							watchKey.SyncStatus = string(v1alpha1.SyncStatusCodeSynced)
							//If revisionHash = "", it means the app can't be found
							revisionHash, err := argoClient.GetCurrentRevisionHash(watchKey.Name)
							if err != nil {
								break
							}
							var healthStatus health.HealthStatusCode
							healthStatus = appInfo.HealthStatus
							if string(healthStatus) != watchKey.HealthStatus {
								watchKey.GeneratedCompletionEvent = false
							}
							watchKey.HealthStatus = string(healthStatus)
							watchKey.SyncStatus = string(appInfo.SyncStatus)
							watchedApplications[mapKey] = watchKey
							resourceStatuses, err := argoClient.GetAppResourcesStatus(watchKey.Name)
							if err != nil {
								log.Printf("Getting operation success failed with error %s\n", err)
								break
							}
							eventInfo = datamodel.MakeApplicationEvent(watchedApplications[mapKey], appInfo, watchKey.HealthStatus, watchKey.SyncStatus, resourceStatuses, revisionHash)
						} else {
							oldSyncStatus := watchKey.SyncStatus
							oldHealthStatus := watchKey.HealthStatus
							watchKey.SyncStatus = string(appInfo.SyncStatus)
							watchKey.HealthStatus = string(appInfo.HealthStatus)

							resourceStatuses, err := argoClient.GetAppResourcesStatus(watchKey.Name)
							if err != nil {
								log.Printf("Getting operation success failed with error %s\n", err)
								break
							}

							createdEvent := false
							if !watchKey.GeneratedCompletionEvent {
								completed, success, revisionHash, err := argoClient.GetOperationSuccess(appInfo.Name)
								if err != nil {
									log.Printf("Getting operation success failed with error %s\n", err)
									break
								}
								if completed {
									healthStatus := string(health.HealthStatusMissing)
									if success {
										healthStatus = string(health.HealthStatusHealthy)
									} else {
										healthStatus = string(health.HealthStatusUnknown)
									}
									eventInfo = datamodel.MakeApplicationEvent(watchedApplications[mapKey], appInfo, healthStatus, watchKey.SyncStatus, resourceStatuses, revisionHash)
									createdEvent = true
								}
							}
							if oldSyncStatus == watchKey.SyncStatus && oldHealthStatus == watchKey.HealthStatus && !createdEvent {
								break
							}
							watchKey.GeneratedCompletionEvent = false
							watchedApplications[mapKey] = watchKey
							//If the application is not in sync, it is difficult to know whether an operation is progressing or not. We don't want to prematurely send the wrong revisionId.
							//They are temporarily set to -1 for now, given that the events generated below are remediation events.
							//TODO: We will have to figure out what they should be and how we can accurately get the revisionId at all times.
							if !createdEvent {
								if watchKey.HealthStatus != oldHealthStatus {
									eventInfo = datamodel.MakeApplicationEvent(watchedApplications[mapKey], appInfo, watchKey.HealthStatus, watchKey.SyncStatus, resourceStatuses, "")
								} else {
									eventInfo = datamodel.MakeApplicationEvent(watchedApplications[mapKey], appInfo, watchKey.SyncStatus, watchKey.SyncStatus, resourceStatuses, "")
								}
							}
						}
						for i := 0; i < HttpRequestRetryLimit; i++ {
							if !watchedApplications[mapKey].GeneratedCompletionEvent && generateEvent(eventInfo, workflowTriggerAddress) {
								log.Printf("Generated Client Completion event for %s", appInfo.Name)
								watchKey.GeneratedCompletionEvent = true
								watchedApplications[mapKey] = watchKey
								break
							}
						}
						break
					}
				}
			} else if watchKey.Type == datamodel.WatchTestKey {
				jobStatus, selector, numPods := kubernetesClient.GetJob(watchKey.Name, watchKey.Namespace)
				log.Printf("Job status succeeded: %d, status failed %d", jobStatus.Succeeded, jobStatus.Failed)
				if numPods == -1 {
					log.Printf("Getting the Job failed")
					continue
				}
				if jobStatus.Succeeded == numPods || jobStatus.Failed >= numPods {
					podLogs, err := kubernetesClient.GetLogs(watchKey.Namespace, selector)
					if err != nil {
						//TODO: This should have a timeout. For example, if it fails 5 times just send the event without the logs
						log.Printf("Pod logs could not be fetched...will try again next time")
						continue
					}
					eventInfo := datamodel.MakeTestEvent(watchKey, jobStatus.Succeeded > 0, podLogs)
					for i := 0; i < HttpRequestRetryLimit; i++ {
						if !watchKey.GeneratedCompletionEvent && generateEvent(eventInfo, workflowTriggerAddress) {
							log.Printf("Generated Test Completion event for %s", watchKey.Name)
							watchKey.GeneratedCompletionEvent = true
							watchedApplications[mapKey] = watchKey
							if !kubernetesClient.Delete(watchKey.Name, watchKey.Namespace, schema.GroupVersionKind{Kind: k8sdriver.JobType}) {
								continue
							}
							deleteKeys = append(deleteKeys, mapKey)
							break
						} else if kubernetesClient.Delete(watchKey.Name, watchKey.Namespace, schema.GroupVersionKind{Kind: k8sdriver.JobType}) {
							deleteKeys = append(deleteKeys, mapKey)
							break
						}
					}
				}
			}
		}
		for _, keyToDelete := range deleteKeys {
			delete(watchedApplications, keyToDelete)
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
		metricVariablePattern, _ := regexp.Compile("([a-zA-Z_-]*)=\"([a-zA-Z0-9:/_.-]*)\"")
		variableMap := make(map[string]string)
		for _, variable := range metricVariablePattern.FindAllStringSubmatch(payloadLine, -1) {
			variableMap[variable[1]] = variable[2]
		}
		//We should be using reflection in the future to assign all these values. For now this will work.
		appInfo := datamodel.ArgoAppMetricInfo{
			DestNamespace: variableMap["dest_namespace"],
			DestServer:    variableMap["dest_server"],
			HealthStatus:  health.HealthStatusCode(variableMap["health_status"]),
			Name:          variableMap["name"],
			Namespace:     variableMap["namespace"],
			Operation:     variableMap["operation"],
			Project:       variableMap["project"],
			Repo:          variableMap["repo"],
			SyncStatus:    v1alpha1.SyncStatusCode(variableMap["sync_status"]),
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

func Start(kubernetesClient k8sdriver.KubernetesClientGetRestricted, argoClient argodriver.ArgoGetRestrictedClient, channel chan string) {
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
	checkForCompletedApplications(kubernetesClient, argoClient, progressionCheckerChannel, metricsServerAddress, workflowTriggerAddress)
}
