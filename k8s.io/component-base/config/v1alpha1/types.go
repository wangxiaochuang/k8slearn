package v1alpha1

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const EndpointsResourceLock = "endpoints"

type LeaderElectionConfiguration struct {
	LeaderElect       *bool           `json:"leaderElect"`
	LeaseDuration     metav1.Duration `json:"leaseDuration"`
	RenewDeadline     metav1.Duration `json:"renewDeadline"`
	RetryPeriod       metav1.Duration `json:"retryPeriod"`
	ResourceLock      string          `json:"resourceLock"`
	ResourceName      string          `json:"resourceName"`
	ResourceNamespace string          `json:"resourceNamespace"`
}

type DebuggingConfiguration struct {
	EnableProfiling           *bool `json:"enableProfiling,omitempty"`
	EnableContentionProfiling *bool `json:"enableContentionProfiling,omitempty"`
}

type ClientConnectionConfiguration struct {
	Kubeconfig         string  `json:"kubeconfig"`
	AcceptContentTypes string  `json:"acceptContentTypes"`
	ContentType        string  `json:"contentType"`
	QPS                float32 `json:"qps"`
	Burst              int32   `json:"burst"`
}

type LoggingConfiguration struct {
	Format         string               `json:"format,omitempty"`
	FlushFrequency time.Duration        `json:"flushFrequency"`
	Verbosity      uint32               `json:"verbosity"`
	VModule        VModuleConfiguration `json:"vmodule,omitempty"`
	Options        FormatOptions        `json:"options,omitempty"`
}

type FormatOptions struct {
	// [Experimental] JSON contains options for logging format "json".
	JSON JSONOptions `json:"json,omitempty"`
}

type JSONOptions struct {
	SplitStream    bool                   `json:"splitStream,omitempty"`
	InfoBufferSize resource.QuantityValue `json:"infoBufferSize,omitempty"`
}

type VModuleConfiguration []VModuleItem

type VModuleItem struct {
	FilePattern string `json:"filePattern"`
	Verbosity   uint32 `json:"verbosity"`
}
