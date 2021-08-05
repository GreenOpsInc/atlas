package argodriver

import (
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
)

func (a ArgoClientDriver) SpecMatches(app1 *v1alpha1.Application, app2 *v1alpha1.Application) bool {
	if app1 == nil || app2 == nil {
		return false
	}
	return app1.Spec.String() == app2.Spec.String()
}
