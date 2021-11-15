package argoworkflows

import (
	"fmt"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"greenops.io/client/progressionchecker/datamodel"
	"log"
)

const (
	LineSeparator string = "\n---"
)

func (a *ArgoWfClientDriver) CheckStatus(watchKey datamodel.WatchKey) datamodel.EventInfo {
	status, err := a.GetWorkflowStatus(watchKey.Name, watchKey.Namespace)
	if err != nil {
		log.Printf("Failed to get the Argo Workflow's status: %s", err)
		return nil
	}

	if status.Phase.Completed() {
		var eventInfo datamodel.EventInfo
		if status.Phase == wfv1.WorkflowSucceeded {
			eventInfo = datamodel.MakeTestEvent(watchKey, true, status.Message)
		} else {
			message := status.Message
			message += LineSeparator
			nodes := status.Nodes.Filter(getFailedNodes)
			for _, v := range nodes {
				message += fmt.Sprintf("Failed node: %s\n%s\n\n", v.DisplayName, v.Message)
			}
			eventInfo = datamodel.MakeTestEvent(watchKey, false, message)
		}
		return eventInfo
	}
	return nil
}

func getFailedNodes(nodeStatus wfv1.NodeStatus) bool {
	return nodeStatus.Phase == wfv1.NodeFailed || nodeStatus.Phase == wfv1.NodeError
}
