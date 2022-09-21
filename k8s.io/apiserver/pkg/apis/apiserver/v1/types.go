package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AdmissionConfiguration struct {
	metav1.TypeMeta `json:",inline"`
	Plugins         []AdmissionPluginConfiguration `json:"plugins"`
}

type AdmissionPluginConfiguration struct {
	Name          string           `json:"name"`
	Path          string           `json:"path"`
	Configuration *runtime.Unknown `json:"configuration"`
}
