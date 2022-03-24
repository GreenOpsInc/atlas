package v1alpha1

import (
	"github.com/greenopsinc/pipeline"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion                = schema.GroupVersion{Group: pipeline.Group, Version: "v1alpha1"}
	ApplicationSchemaGroupVersionKind = schema.GroupVersionKind{Group: pipeline.Group, Version: "v1alpha1", Kind: pipeline.ApplicationKind}
	AppProjectSchemaGroupVersionKind  = schema.GroupVersionKind{Group: pipeline.Group, Version: "v1alpha1", Kind: pipeline.AppProjectKind}
)

// Resource takes an unqualified resource and returns a Group-qualified GroupResource.
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion, &Pipeline{}, &PipelineList{})
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
