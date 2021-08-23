package argodriver

import (
	"errors"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"log"
	"strings"
)

const (
	resourceFieldDelimiter              = ":"
	resourceFieldCount                  = 4
)

func (a ArgoClientDriver) SpecMatches(app1 *v1alpha1.Application, app2 *v1alpha1.Application) bool {
	if app1 == nil || app2 == nil {
		return false
	}
	return app1.Spec.String() == app2.Spec.String()
}

func (a ArgoClientDriver) ParseSelectedResources(resources []string) ([]v1alpha1.SyncOperationResource, error) {
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
