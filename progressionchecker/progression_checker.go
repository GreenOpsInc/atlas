package progressionchecker

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	EndTransactionMarker         string = "__ATLAS_END_TRANSACTION_MARKER__"
	Missing                      string = "Missing"
	NotHealthy                   string = "NotHealthy"
	Healthy                      string = "Healthy"
	ProgressionChannelBufferSize int    = 100
	DefaultMetricsServerAddress  string = "http://argocd-metrics.argocd.svc.cluster.local:8082/metrics"
	MetricsServerEnvVar          string = "ARGOCD_METRICS_SERVER_ADDR"
	MetricsRetryLimit            int    = 3
)

func listenForApplicationsToWatch(inputChannel chan string, outputChannel chan string) {
	var applications []string
	for {
		for {
			newApplicationName := <-inputChannel
			if newApplicationName == EndTransactionMarker {
				break
			}
			applications = append(applications, newApplicationName)
		}
		for len(applications) > 0 {
			outputChannel <- applications[0]
			applications = applications[1:]
		}
	}
}

func checkForCompletedApplications(channel chan string, metricsServerAddress string) {
	watchedApplications := make(map[string]string)
	for {
		//Update watched applications list with new entries
		for {
			doneReading := false
			select {
			case newApplicationName := <-channel:
				if _, ok := watchedApplications[newApplicationName]; !ok {
					watchedApplications[newApplicationName] = Missing
					log.Printf("Added Argo CD application %s to watched list\n", newApplicationName)
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
		payload := ""
		for i := 0; i < MetricsRetryLimit; i++ {
			payload = getMetrics(metricsServerAddress)
			if payload != "" {
				break
			}
		}

		duration := 10 * time.Second
		time.Sleep(duration)
	}
}

func getMetrics(metricsServerAddress string) string {
	resp, err := http.Get(metricsServerAddress)
	if err != nil {
		log.Printf("Error querying metrics server. Error was %s\n", err)
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading in response body from metrics server. Error was %s\n", err)
		return ""
	}
	return string(body)
}

func Start(channel chan string) {
	metricsServerAddress := os.Getenv(MetricsServerEnvVar)
	if metricsServerAddress == "" {
		metricsServerAddress = DefaultMetricsServerAddress
	}
	progressionCheckerChannel := make(chan string, ProgressionChannelBufferSize)
	//TODO: Add initial cache creation
	go listenForApplicationsToWatch(channel, progressionCheckerChannel)
	checkForCompletedApplications(progressionCheckerChannel, metricsServerAddress)
}
