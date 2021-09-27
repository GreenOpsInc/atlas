package k8sdriver

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilpointer "k8s.io/utils/pointer"
)

func (k KubernetesClientDriver) createConfigMap(name string, namespace string, entries map[string]string) *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{Kind: ConfigMapType, APIVersion: v1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Immutable:  utilpointer.BoolPtr(true),
		Data:       entries,
	}
}
