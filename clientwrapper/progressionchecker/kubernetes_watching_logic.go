package progressionchecker

import (
	"greenops.io/client/k8sdriver"
	"greenops.io/client/progressionchecker/datamodel"
	"log"
)

func CheckKubernetesStatus(watchKey datamodel.WatchKey, kubernetesClient k8sdriver.KubernetesClientGetRestricted) datamodel.EventInfo {
	var eventInfo datamodel.EventInfo
	jobStatus, selector, numPods := kubernetesClient.GetJob(watchKey.Name, watchKey.Namespace)
	log.Printf("Job status succeeded: %d, status failed %d", jobStatus.Succeeded, jobStatus.Failed)
	if numPods == -1 {
		log.Printf("Getting the Job failed")
		return nil
	}
	if jobStatus.Succeeded == numPods || jobStatus.Failed >= numPods {
		podLogs, err := kubernetesClient.GetLogs(watchKey.Namespace, selector)
		if err != nil {
			//TODO: This should have a timeout. For example, if it fails 5 times just send the event without the logs
			log.Printf("Pod logs could not be fetched...will try again next time")
			return nil
		}
		eventInfo = datamodel.MakeTestEvent(watchKey, jobStatus.Succeeded > 0, podLogs)
	}
	return eventInfo
}
