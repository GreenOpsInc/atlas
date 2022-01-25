package argodriver

import (
	"errors"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"log"
	"reflect"
	"strings"
)

const (
	resourceFieldDelimiter = ":"
	resourceFieldCount     = 4
)

func (a *ArgoClientDriver) SpecMatches(app1 *v1alpha1.Application, app2 *v1alpha1.Application) bool {
	if app1 == nil || app2 == nil {
		return false
	}
	return app1.Name == app2.Name &&
		app1.Namespace == app2.Namespace &&
		app1.ClusterName == app2.ClusterName &&
		reflect.DeepEqual(app1.Annotations, app2.Annotations) &&
		reflect.DeepEqual(app1.Labels, app2.Labels) &&
		app1.Spec.String() == app2.Spec.String()
}

func (a *ArgoClientDriver) ParseSelectedResources(resources []string) ([]v1alpha1.SyncOperationResource, error) {
	var selectedResources []v1alpha1.SyncOperationResource
	if resources != nil {
		selectedResources = []v1alpha1.SyncOperationResource{}
		for _, r := range resources {
			fields := strings.Split(r, resourceFieldDelimiter)
			if len(fields) != resourceFieldCount {
				log.Printf("Resource should have GROUP%sKIND%sNAME%sNAMESPACE, but instead got: %s", resourceFieldDelimiter, resourceFieldDelimiter, resourceFieldDelimiter, r)
				return selectedResources, errors.New("could not strip resources correctly")
			}
			rsrc := v1alpha1.SyncOperationResource{
				Group:     fields[0],
				Kind:      fields[1],
				Name:      fields[2],
				Namespace: fields[3],
			}
			selectedResources = append(selectedResources, rsrc)
		}
	}
	return selectedResources, nil
}

func containsHealth(s []health.HealthStatusCode, e health.HealthStatusCode) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func containsSync(s []v1alpha1.SyncStatusCode, e v1alpha1.SyncStatusCode) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
