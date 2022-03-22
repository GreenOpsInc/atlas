package progressionchecker

import (
	"encoding/json"
	"errors"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"greenops.io/client/api/generation"
	"greenops.io/client/argodriver"
	"greenops.io/client/k8sdriver"
	"greenops.io/client/plugins"
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
	MetricsServerEnvVar   string = "ARGOCD_METRICS_SERVER_ADDR"
	HttpRequestRetryLimit int    = 3
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

func checkForCompletedApplications(watchedApplications map[string]datamodel.WatchKey, kubernetesClient k8sdriver.KubernetesClientGetRestricted, argoClient argodriver.ArgoGetRestrictedClient, eventGenerationApi generation.EventGenerationApi, pluginList plugins.Plugins, channel chan string, metricsServerAddress string) (recoveredWatchedApplications map[string]datamodel.WatchKey) {
	var mapKey string
	var watchKey datamodel.WatchKey
	defer func() {
		if err := recover(); err != nil {
			var stronglyTypedError error
			switch err.(type) {
			case error:
				stronglyTypedError = err.(error)
			default:
				stronglyTypedError = errors.New(err.(string))
			}
			log.Printf("Progression checker exiting with error %s", stronglyTypedError.Error())
			failureEvent := datamodel.MakeFailureEventEventRaw(watchKey.OrgName, watchKey.TeamName, watchKey.PipelineName, watchKey.PipelineUvn, watchKey.StepName, stronglyTypedError.Error())
			for i := 0; i < HttpRequestRetryLimit; i++ {
				if eventGenerationApi.GenerateEvent(failureEvent) {
					delete(watchedApplications, mapKey)
					break
				}
			}
		}
		//recoveredWatchedApplications is the named return type
		//In case of an error, this recovery function return the list of watched applications by setting recoveredWatchedApplications
		recoveredWatchedApplications = watchedApplications
	}()
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
				if err := PersistWatchedKeys(kubernetesClient, watchedApplications); err == nil {
					break
				}
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
		for mapKey, watchKey = range watchedApplications {
			if watchKey.Type == datamodel.WatchArgoApplicationKey {
				for _, appInfo := range argoAppInfo {
					if watchKey.Name == appInfo.Name && watchKey.Namespace == appInfo.Namespace {
						newWatchKey, eventInfo := CheckArgoCdStatus(watchKey, appInfo, argoClient)
						if newWatchKey == GetEmptyWatchKey() && eventInfo == nil {
							break
						}
						watchedApplications[mapKey] = newWatchKey
						for i := 0; i < HttpRequestRetryLimit; i++ {
							if !watchedApplications[mapKey].GeneratedCompletionEvent && eventGenerationApi.GenerateEvent(eventInfo) {
								log.Printf("Generated Client Completion event for %s", appInfo.Name)
								newWatchKey.GeneratedCompletionEvent = true
								watchedApplications[mapKey] = newWatchKey
								break
							}
						}
						break
					}
				}
			} else if watchKey.Type == datamodel.WatchTestKey {
				for i := 0; i < HttpRequestRetryLimit; i++ {
					eventInfo := CheckKubernetesStatus(watchKey, kubernetesClient)
					if eventInfo == nil {
						continue
					}
					if !watchKey.GeneratedCompletionEvent && eventGenerationApi.GenerateEvent(eventInfo) {
						log.Printf("Generated Test Completion event for %s", watchKey.Name)
						watchKey.GeneratedCompletionEvent = true
						watchedApplications[mapKey] = watchKey
						if TaskCleanup(watchKey, kubernetesClient) != nil {
							continue
						}
						deleteKeys = append(deleteKeys, mapKey)
						break
					} else if TaskCleanup(watchKey, kubernetesClient) == nil {
						deleteKeys = append(deleteKeys, mapKey)
						break
					}
				}
			} else if watchKey.Type == datamodel.WatchArgoWorkflowKey {
				for i := 0; i < HttpRequestRetryLimit; i++ {
					plugin, err := pluginList.GetPlugin(plugins.PluginType(datamodel.WatchArgoWorkflowKey))
					if err != nil {
						log.Printf("Could not fetch plugin correctly: %s", err)
						continue
					}
					eventInfo := plugin.PluginObject.CheckStatus(watchKey)
					if eventInfo == nil {
						continue
					}
					if !watchKey.GeneratedCompletionEvent && eventGenerationApi.GenerateEvent(eventInfo) {
						log.Printf("Generated Test Completion event for %s", watchKey.Name)
						watchKey.GeneratedCompletionEvent = true
						watchedApplications[mapKey] = watchKey
						if plugin.PluginObject.Cleanup(watchKey) != nil {
							continue
						}
						deleteKeys = append(deleteKeys, mapKey)
						break
					} else if plugin.PluginObject.Cleanup(watchKey) == nil {
						deleteKeys = append(deleteKeys, mapKey)
						break
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

func Start(kubernetesClient k8sdriver.KubernetesClientGetRestricted, argoClient argodriver.ArgoGetRestrictedClient, eventGenerationApi generation.EventGenerationApi, pluginList plugins.Plugins, channel chan string) {
	metricsServerAddress := os.Getenv(MetricsServerEnvVar)
	if metricsServerAddress == "" {
		metricsServerAddress = DefaultMetricsServerAddress
	}
	progressionCheckerChannel := make(chan string, ProgressionChannelBufferSize)
	go listenForApplicationsToWatch(channel, progressionCheckerChannel)
	watchedApplications := RetrieveWatchedKeys(kubernetesClient)
	for {
		log.Printf("Progression checker starting...")
		watchedApplications = checkForCompletedApplications(watchedApplications, kubernetesClient, argoClient, eventGenerationApi, pluginList, progressionCheckerChannel, metricsServerAddress)
	}
}
