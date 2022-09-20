package audit

import (
	authnv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	HeaderAuditID = "Audit-ID"
)

type Level string

const (
	LevelNone            Level = "None"
	LevelMetadata        Level = "Metadata"
	LevelRequest         Level = "Request"
	LevelRequestResponse Level = "RequestResponse"
)

type Stage string

const (
	StageRequestReceived  Stage = "RequestReceived"
	StageResponseStarted  Stage = "ResponseStarted"
	StageResponseComplete Stage = "ResponseComplete"
	StagePanic            Stage = "Panic"
)

type Event struct {
	metav1.TypeMeta

	Level Level

	AuditID types.UID
	Stage   Stage

	RequestURI               string
	Verb                     string
	User                     authnv1.UserInfo
	ImpersonatedUser         *authnv1.UserInfo
	SourceIPs                []string
	UserAgent                string
	ObjectRef                *ObjectReference
	ResponseStatus           *metav1.Status
	RequestObject            *runtime.Unknown
	ResponseObject           *runtime.Unknown
	RequestReceivedTimestamp metav1.MicroTime
	StageTimestamp           metav1.MicroTime
	Annotations              map[string]string
}

type EventList struct {
	metav1.TypeMeta
	// +optional
	metav1.ListMeta

	Items []Event
}

type Policy struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	Rules             []PolicyRule
	OmitStages        []Stage
	OmitManagedFields bool
}

type PolicyList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []Policy
}

type PolicyRule struct {
	Level             Level
	Users             []string
	UserGroups        []string
	Verbs             []string
	Resources         []GroupResources
	Namespaces        []string
	NonResourceURLs   []string
	OmitStages        []Stage
	OmitManagedFields *bool
}

type GroupResources struct {
	Group         string
	Resources     []string
	ResourceNames []string
}

type ObjectReference struct {
	Resource        string
	Namespace       string
	Name            string
	UID             types.UID
	APIGroup        string
	APIVersion      string
	ResourceVersion string
	Subresource     string
}
