package progressionchecker

import (
	"bytes"
	"encoding/json"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	EndTransactionMarker          string = "__ATLAS_END_TRANSACTION_MARKER__"
	Missing                       string = "Missing"
	NotHealthy                    string = "NotHealthy"
	Healthy                       string = "Healthy"
	ProgressionChannelBufferSize  int    = 100
	DefaultMetricsServerAddress   string = "http://argocd-metrics.argocd.svc.cluster.local:8082/metrics"
	DefaultWorkflowTriggerAddress string = "http://atlasworkflowtrigger.default.svc.cluster.local"
	MetricsServerEnvVar           string = "ARGOCD_METRICS_SERVER_ADDR"
	WorkflowTriggerEnvVar         string = "WORKFLOW_TRIGGER_SERVER_ADDR"
	HttpRequestRetryLimit         int    = 3
)

type ArgoAppMetricInfo struct {
	DestNamespace string `tag:"dest_namespace"`
	DestServer    string `tag:"dest_server"`
	HealthStatus  string `tag:"health_status"`
	Name          string `tag:"name"`
	Namespace     string `tag:"namespace"`
	Operation     string `tag:"operation"`
	Project       string `tag:"project"`
	Repo          string `tag:"repo"`
	SyncStatus    string `tag:"sync_status"`
}

type EventInfo struct {
	HealthStatus string
	PipelineName string
	StepName     string
	argoName     string
	Operation    string
	Project      string
	Repo         string
}

type WatchKey struct {
	PipelineName string
	StepName     string
	AppName      string
	//You can't have two Argo apps with the same name in the same namespace, this makes sure there are no collisions
	Namespace string
}

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

func checkForCompletedApplications(channel chan string, metricsServerAddress string, workflowTriggerAddress string) {
	watchedApplications := make(map[WatchKey]string)
	for {
		//Update watched applications list with new entries
		for {
			doneReading := false
			select {
			case newKey := <-channel:
				newWatchKey := readKeyAsWatchKey(newKey)
				if _, ok := watchedApplications[newWatchKey]; !ok {
					watchedApplications[newWatchKey] = Missing
					log.Printf("Added Argo CD application %s to watched list\n", newWatchKey)
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
		var argoAppInfo []ArgoAppMetricInfo
		for i := 0; i < HttpRequestRetryLimit; i++ {
			argoAppInfo = getUnmarshalledMetrics(metricsServerAddress)
			if argoAppInfo != nil {
				break
			}
		}

		for _, appInfo := range argoAppInfo {
			key := WatchKey{}
			for mapKey, _ := range watchedApplications {
				if mapKey.AppName == appInfo.Name && mapKey.Namespace == appInfo.Namespace {
					key = mapKey
					break
				}
			}
			if key.PipelineName == "" {
				log.Printf("The Argo application named %s was never marked as \"watched\"", appInfo.Name)
				continue
			}
			if appInfo.HealthStatus == string(health.HealthStatusHealthy) && appInfo.SyncStatus == string(v1alpha1.SyncStatusCodeSynced) {
				watchedApplications[key] = Healthy
				for i := 0; i < HttpRequestRetryLimit; i++ {
					eventInfo := EventInfo{
						HealthStatus: Healthy,
						PipelineName: key.PipelineName,
						StepName:     key.StepName,
						argoName:     appInfo.Name,
						Operation:    appInfo.Operation,
						Project:      appInfo.Project,
						Repo:         appInfo.Repo,
					}
					generateEvent(eventInfo, workflowTriggerAddress)
				}
			} else {
				watchedApplications[key] = NotHealthy
			}
		}

		duration := 10 * time.Second
		time.Sleep(duration)
	}
}

func unmarshall(payload *string) []ArgoAppMetricInfo {
	metricTypePattern, _ := regexp.Compile("(argocd_app_info{.*} [0-9]*)+")
	relevantPayload := metricTypePattern.FindAllString(*payload, -1)
	var argoAppInfo []ArgoAppMetricInfo
	for _, payloadLine := range relevantPayload {
		metricVariablePattern, _ := regexp.Compile("([a-zA-Z_-]*)=\"([a-zA-Z:/_.-]*)\"")
		variableMap := make(map[string]string)
		for _, variable := range metricVariablePattern.FindAllStringSubmatch(payloadLine, -1) {
			variableMap[variable[1]] = variable[2]
		}
		//We should be using reflection in the future to assign all these values. For now this will work.
		appInfo := ArgoAppMetricInfo{
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

func getUnmarshalledMetrics(metricsServerAddress string) []ArgoAppMetricInfo {
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

func generateEvent(eventInfo EventInfo, workflowTriggerAddress string) bool {
	data, err := json.Marshal(eventInfo)
	if err != nil {
		return false
	}
	resp, err := http.Post(workflowTriggerAddress, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error generating progression event. Error was %s\n", err)
		return false
	}
	defer resp.Body.Close()
	return strings.HasPrefix(resp.Status, "2")
}

func readKeyAsWatchKey(key string) WatchKey {
	splitKey := strings.Split(key, "-")
	return WatchKey{
		PipelineName: splitKey[0],
		StepName:     splitKey[1],
		AppName:      splitKey[2],
		Namespace:    splitKey[3],
	}
}

func (watchKey WatchKey) WriteKeyAsString() string {
	return strings.Join([]string{watchKey.PipelineName, watchKey.StepName, watchKey.AppName, watchKey.Namespace}, "-")
}

func Start(channel chan string) {
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
	checkForCompletedApplications(progressionCheckerChannel, metricsServerAddress, workflowTriggerAddress)
}
