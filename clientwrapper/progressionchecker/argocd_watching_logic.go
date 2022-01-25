package progressionchecker

import (
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"greenops.io/client/argodriver"
	"greenops.io/client/progressionchecker/datamodel"
	"log"
)

const (
	emptyWatchKeyName string = "EMPTYWATCHKEY"
)

func CheckArgoCdStatus(watchKey datamodel.WatchKey, appInfo datamodel.ArgoAppMetricInfo, argoClient argodriver.ArgoGetRestrictedClient) (datamodel.WatchKey, datamodel.EventInfo) {
	if appInfo.HealthStatus == health.HealthStatusProgressing {
		return GetEmptyWatchKey(), nil
	} else if appInfo.SyncStatus == v1alpha1.SyncStatusCodeSynced {
		return CheckArgoCdSyncedStatus(watchKey, appInfo, argoClient)
	} else {
		return CheckArgoCdNonSyncedStatus(watchKey, appInfo, argoClient)
	}
}

func CheckArgoCdSyncedStatus(watchKey datamodel.WatchKey, appInfo datamodel.ArgoAppMetricInfo, argoClient argodriver.ArgoGetRestrictedClient) (datamodel.WatchKey, datamodel.EventInfo) {
	var eventInfo datamodel.EventInfo
	watchKey.SyncStatus = string(v1alpha1.SyncStatusCodeSynced)
	//If revisionHash = "", it means the app can't be found
	revisionHash, err := argoClient.GetCurrentRevisionHash(watchKey.Name)
	if err != nil {
		return GetEmptyWatchKey(), nil
	}
	var healthStatus health.HealthStatusCode
	healthStatus = appInfo.HealthStatus
	if string(healthStatus) != watchKey.HealthStatus {
		watchKey.GeneratedCompletionEvent = false
	}
	watchKey.HealthStatus = string(healthStatus)
	watchKey.SyncStatus = string(appInfo.SyncStatus)
	//watchedApplications[mapKey] = watchKey
	resourceStatuses, err := argoClient.GetAppResourcesStatus(watchKey.Name)
	if err != nil {
		log.Printf("Getting operation success failed with error %s\n", err)
		return GetEmptyWatchKey(), nil
	}
	eventInfo = datamodel.MakeApplicationEvent(watchKey, appInfo, watchKey.HealthStatus, watchKey.SyncStatus, resourceStatuses, revisionHash)
	return watchKey, eventInfo
}

//func CheckArgoSelectiveSyncStatus(watchKey datamodel.WatchKey, appInfo datamodel.ArgoAppMetricInfo, argoClient argodriver.ArgoGetRestrictedClient) (datamodel.WatchKey, datamodel.EventInfo) {
//	var eventInfo datamodel.EventInfo
//	appResources, err := argoClient.GetAppResourcesStatus(watchKey.Name)
//	if err != nil {
//		return GetEmptyWatchKey(), nil
//	}
//	for _, appResource := range *watchKey.Resources {
//		for _, appResourceSrc := range appResources {
//			if ResourceMatchesIgnoreStatus(appResource, appResourceSrc) {
//				appResourceSrc.HealthStatus ==
//			}
//		}
//	}
//}

func CheckArgoCdNonSyncedStatus(watchKey datamodel.WatchKey, appInfo datamodel.ArgoAppMetricInfo, argoClient argodriver.ArgoGetRestrictedClient) (datamodel.WatchKey, datamodel.EventInfo) {
	var eventInfo datamodel.EventInfo
	oldSyncStatus := watchKey.SyncStatus
	oldHealthStatus := watchKey.HealthStatus
	watchKey.SyncStatus = string(appInfo.SyncStatus)
	watchKey.HealthStatus = string(appInfo.HealthStatus)

	resourceStatuses, err := argoClient.GetAppResourcesStatus(watchKey.Name)
	if err != nil {
		log.Printf("Getting operation success failed with error %s\n", err)
		return GetEmptyWatchKey(), nil
	}

	createdEvent := false
	if !watchKey.GeneratedCompletionEvent {
		completed, success, revisionHash, err := argoClient.GetOperationSuccess(appInfo.Name)
		if err != nil {
			log.Printf("Getting operation success failed with error %s\n", err)
			return GetEmptyWatchKey(), nil
		}
		if completed {
			healthStatus := string(health.HealthStatusMissing)
			if success {
				healthStatus = string(health.HealthStatusHealthy)
			} else {
				healthStatus = string(health.HealthStatusUnknown)
			}
			eventInfo = datamodel.MakeApplicationEvent(watchKey, appInfo, healthStatus, watchKey.SyncStatus, resourceStatuses, revisionHash)
			createdEvent = true
		} else {
			//Status is either terminating or running (not complete), so this iteration should be skipped and revisted
			return GetEmptyWatchKey(), nil
		}
	}
	if oldSyncStatus == watchKey.SyncStatus && oldHealthStatus == watchKey.HealthStatus && !createdEvent {
		return GetEmptyWatchKey(), nil
	}
	watchKey.GeneratedCompletionEvent = false
	//If the application is not in sync, it is difficult to know whether an operation is progressing or not. We don't want to prematurely send the wrong revisionId.
	//They are temporarily set to -1 for now, given that the events generated below are remediation events.
	//TODO: We will have to figure out what they should be and how we can accurately get the revisionId at all times.
	if !createdEvent {
		if watchKey.HealthStatus != oldHealthStatus {
			eventInfo = datamodel.MakeApplicationEvent(watchKey, appInfo, watchKey.HealthStatus, watchKey.SyncStatus, resourceStatuses, "")
		} else {
			eventInfo = datamodel.MakeApplicationEvent(watchKey, appInfo, watchKey.SyncStatus, watchKey.SyncStatus, resourceStatuses, "")
		}
	}
	return watchKey, eventInfo
}

func GetEmptyWatchKey() datamodel.WatchKey {
	return datamodel.WatchKey{Name: emptyWatchKeyName}
}

func ResourceMatchesIgnoreStatus(resource1 datamodel.ResourceStatus, resource2 datamodel.ResourceStatus) bool {
	return resource1.ResourceName == resource2.ResourceName && resource1.ResourceNamespace == resource2.ResourceNamespace && resource1.Group == resource2.Group && resource1.Kind == resource2.Kind && resource1.Version == resource2.Version
}
