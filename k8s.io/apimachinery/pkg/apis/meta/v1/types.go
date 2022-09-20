package v1

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type TypeMeta struct {
	Kind       string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,2,opt,name=apiVersion"`
}

type ListMeta struct {
	SelfLink           string `json:"selfLink,omitempty" protobuf:"bytes,1,opt,name=selfLink"`
	ResourceVersion    string `json:"resourceVersion,omitempty" protobuf:"bytes,2,opt,name=resourceVersion"`
	Continue           string `json:"continue,omitempty" protobuf:"bytes,3,opt,name=continue"`
	RemainingItemCount *int64 `json:"remainingItemCount,omitempty" protobuf:"bytes,4,opt,name=remainingItemCount"`
}

const (
	ObjectNameField = "metadata.name"
)

const (
	FinalizerOrphanDependents = "orphan"
	FinalizerDeleteDependents = "foregroundDeletion"
)

type ObjectMeta struct {
	Name                       string               `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	GenerateName               string               `json:"generateName,omitempty" protobuf:"bytes,2,opt,name=generateName"`
	Namespace                  string               `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`
	SelfLink                   string               `json:"selfLink,omitempty" protobuf:"bytes,4,opt,name=selfLink"`
	UID                        types.UID            `json:"uid,omitempty" protobuf:"bytes,5,opt,name=uid,casttype=k8s.io/kubernetes/pkg/types.UID"`
	ResourceVersion            string               `json:"resourceVersion,omitempty" protobuf:"bytes,6,opt,name=resourceVersion"`
	Generation                 int64                `json:"generation,omitempty" protobuf:"varint,7,opt,name=generation"`
	CreationTimestamp          Time                 `json:"creationTimestamp,omitempty" protobuf:"bytes,8,opt,name=creationTimestamp"`
	DeletionTimestamp          *Time                `json:"deletionTimestamp,omitempty" protobuf:"bytes,9,opt,name=deletionTimestamp"`
	DeletionGracePeriodSeconds *int64               `json:"deletionGracePeriodSeconds,omitempty" protobuf:"varint,10,opt,name=deletionGracePeriodSeconds"`
	Labels                     map[string]string    `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`
	Annotations                map[string]string    `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`
	OwnerReferences            []OwnerReference     `json:"ownerReferences,omitempty" patchStrategy:"merge" patchMergeKey:"uid" protobuf:"bytes,13,rep,name=ownerReferences"`
	Finalizers                 []string             `json:"finalizers,omitempty" patchStrategy:"merge" protobuf:"bytes,14,rep,name=finalizers"`
	ZZZ_DeprecatedClusterName  string               `json:"clusterName,omitempty" protobuf:"bytes,15,opt,name=clusterName"`
	ManagedFields              []ManagedFieldsEntry `json:"managedFields,omitempty" protobuf:"bytes,17,rep,name=managedFields"`
}

const (
	NamespaceDefault = "default"
	NamespaceAll     = ""
	NamespaceNone    = ""
	NamespaceSystem  = "kube-system"
	NamespacePublic  = "kube-public"
)

type OwnerReference struct {
	APIVersion string    `json:"apiVersion" protobuf:"bytes,5,opt,name=apiVersion"`
	Kind       string    `json:"kind" protobuf:"bytes,1,opt,name=kind"`
	Name       string    `json:"name" protobuf:"bytes,3,opt,name=name"`
	UID        types.UID `json:"uid" protobuf:"bytes,4,opt,name=uid,casttype=k8s.io/apimachinery/pkg/types.UID"`
	Controller *bool     `json:"controller,omitempty" protobuf:"varint,6,opt,name=controller"`

	BlockOwnerDeletion *bool `json:"blockOwnerDeletion,omitempty" protobuf:"varint,7,opt,name=blockOwnerDeletion"`
}

type ListOptions struct {
	TypeMeta             `json:",inline"`
	LabelSelector        string               `json:"labelSelector,omitempty" protobuf:"bytes,1,opt,name=labelSelector"`
	FieldSelector        string               `json:"fieldSelector,omitempty" protobuf:"bytes,2,opt,name=fieldSelector"`
	Watch                bool                 `json:"watch,omitempty" protobuf:"varint,3,opt,name=watch"`
	AllowWatchBookmarks  bool                 `json:"allowWatchBookmarks,omitempty" protobuf:"varint,9,opt,name=allowWatchBookmarks"`
	ResourceVersion      string               `json:"resourceVersion,omitempty" protobuf:"bytes,4,opt,name=resourceVersion"`
	ResourceVersionMatch ResourceVersionMatch `json:"resourceVersionMatch,omitempty" protobuf:"bytes,10,opt,name=resourceVersionMatch,casttype=ResourceVersionMatch"`
	TimeoutSeconds       *int64               `json:"timeoutSeconds,omitempty" protobuf:"varint,5,opt,name=timeoutSeconds"`
	Limit                int64                `json:"limit,omitempty" protobuf:"varint,7,opt,name=limit"`
	Continue             string               `json:"continue,omitempty" protobuf:"bytes,8,opt,name=continue"`
}

type ResourceVersionMatch string

const (
	ResourceVersionMatchNotOlderThan ResourceVersionMatch = "NotOlderThan"
	ResourceVersionMatchExact        ResourceVersionMatch = "Exact"
)

type GetOptions struct {
	TypeMeta        `json:",inline"`
	ResourceVersion string `json:"resourceVersion,omitempty" protobuf:"bytes,1,opt,name=resourceVersion"`
}

type DeletionPropagation string

const (
	DeletePropagationOrphan     DeletionPropagation = "Orphan"
	DeletePropagationBackground DeletionPropagation = "Background"
	DeletePropagationForeground DeletionPropagation = "Foreground"
)

const (
	DryRunAll = "All"
)

type DeleteOptions struct {
	TypeMeta           `json:",inline"`
	GracePeriodSeconds *int64               `json:"gracePeriodSeconds,omitempty" protobuf:"varint,1,opt,name=gracePeriodSeconds"`
	Preconditions      *Preconditions       `json:"preconditions,omitempty" protobuf:"bytes,2,opt,name=preconditions"`
	OrphanDependents   *bool                `json:"orphanDependents,omitempty" protobuf:"varint,3,opt,name=orphanDependents"`
	PropagationPolicy  *DeletionPropagation `json:"propagationPolicy,omitempty" protobuf:"varint,4,opt,name=propagationPolicy"`
	DryRun             []string             `json:"dryRun,omitempty" protobuf:"bytes,5,rep,name=dryRun"`
}

const (
	FieldValidationIgnore = "Ignore"
	FieldValidationWarn   = "Warn"
	FieldValidationStrict = "Strict"
)

type CreateOptions struct {
	TypeMeta        `json:",inline"`
	DryRun          []string `json:"dryRun,omitempty" protobuf:"bytes,1,rep,name=dryRun"`
	FieldManager    string   `json:"fieldManager,omitempty" protobuf:"bytes,3,name=fieldManager"`
	FieldValidation string   `json:"fieldValidation,omitempty" protobuf:"bytes,4,name=fieldValidation"`
}

type PatchOptions struct {
	TypeMeta        `json:",inline"`
	DryRun          []string `json:"dryRun,omitempty" protobuf:"bytes,1,rep,name=dryRun"`
	Force           *bool    `json:"force,omitempty" protobuf:"varint,2,opt,name=force"`
	FieldManager    string   `json:"fieldManager,omitempty" protobuf:"bytes,3,name=fieldManager"`
	FieldValidation string   `json:"fieldValidation,omitempty" protobuf:"bytes,4,name=fieldValidation"`
}

type ApplyOptions struct {
	TypeMeta     `json:",inline"`
	DryRun       []string `json:"dryRun,omitempty" protobuf:"bytes,1,rep,name=dryRun"`
	Force        bool     `json:"force" protobuf:"varint,2,opt,name=force"`
	FieldManager string   `json:"fieldManager" protobuf:"bytes,3,name=fieldManager"`
}

func (o ApplyOptions) ToPatchOptions() PatchOptions {
	return PatchOptions{DryRun: o.DryRun, Force: &o.Force, FieldManager: o.FieldManager}
}

type UpdateOptions struct {
	TypeMeta        `json:",inline"`
	DryRun          []string `json:"dryRun,omitempty" protobuf:"bytes,1,rep,name=dryRun"`
	FieldManager    string   `json:"fieldManager,omitempty" protobuf:"bytes,2,name=fieldManager"`
	FieldValidation string   `json:"fieldValidation,omitempty" protobuf:"bytes,3,name=fieldValidation"`
}

type Preconditions struct {
	UID             *types.UID `json:"uid,omitempty" protobuf:"bytes,1,opt,name=uid,casttype=k8s.io/apimachinery/pkg/types.UID"`
	ResourceVersion *string    `json:"resourceVersion,omitempty" protobuf:"bytes,2,opt,name=resourceVersion"`
}

// p715
type Status struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Status   string         `json:"status,omitempty" protobuf:"bytes,2,opt,name=status"`
	Message  string         `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
	Reason   StatusReason   `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason,casttype=StatusReason"`
	Details  *StatusDetails `json:"details,omitempty" protobuf:"bytes,5,opt,name=details"`
	Code     int32          `json:"code,omitempty" protobuf:"varint,6,opt,name=code"`
}

// p753
type StatusDetails struct {
	Name              string        `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	Group             string        `json:"group,omitempty" protobuf:"bytes,2,opt,name=group"`
	Kind              string        `json:"kind,omitempty" protobuf:"bytes,3,opt,name=kind"`
	UID               types.UID     `json:"uid,omitempty" protobuf:"bytes,6,opt,name=uid,casttype=k8s.io/apimachinery/pkg/types.UID"`
	Causes            []StatusCause `json:"causes,omitempty" protobuf:"bytes,4,rep,name=causes"`
	RetryAfterSeconds int32         `json:"retryAfterSeconds,omitempty" protobuf:"varint,5,opt,name=retryAfterSeconds"`
}

// p783
const (
	StatusSuccess = "Success"
	StatusFailure = "Failure"
)

type StatusReason string

const (
	StatusReasonUnknown               StatusReason = ""
	StatusReasonUnauthorized          StatusReason = "Unauthorized"
	StatusReasonForbidden             StatusReason = "Forbidden"
	StatusReasonNotFound              StatusReason = "NotFound"
	StatusReasonAlreadyExists         StatusReason = "AlreadyExists"
	StatusReasonConflict              StatusReason = "Conflict"
	StatusReasonGone                  StatusReason = "Gone"
	StatusReasonInvalid               StatusReason = "Invalid"
	StatusReasonServerTimeout         StatusReason = "ServerTimeout"
	StatusReasonTimeout               StatusReason = "Timeout"
	StatusReasonTooManyRequests       StatusReason = "TooManyRequests"
	StatusReasonBadRequest            StatusReason = "BadRequest"
	StatusReasonMethodNotAllowed      StatusReason = "MethodNotAllowed"
	StatusReasonNotAcceptable         StatusReason = "NotAcceptable"
	StatusReasonRequestEntityTooLarge StatusReason = "RequestEntityTooLarge"
	StatusReasonUnsupportedMediaType  StatusReason = "UnsupportedMediaType"
	StatusReasonInternalError         StatusReason = "InternalError"
	StatusReasonExpired               StatusReason = "Expired"
	StatusReasonServiceUnavailable    StatusReason = "ServiceUnavailable"
)

// p942
type StatusCause struct {
	Type    CauseType `json:"reason,omitempty" protobuf:"bytes,1,opt,name=reason,casttype=CauseType"`
	Message string    `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	Field   string    `json:"field,omitempty" protobuf:"bytes,3,opt,name=field"`
}

type CauseType string

const (
	CauseTypeFieldValueNotFound       CauseType = "FieldValueNotFound"
	CauseTypeFieldValueRequired       CauseType = "FieldValueRequired"
	CauseTypeFieldValueDuplicate      CauseType = "FieldValueDuplicate"
	CauseTypeFieldValueInvalid        CauseType = "FieldValueInvalid"
	CauseTypeFieldValueNotSupported   CauseType = "FieldValueNotSupported"
	CauseTypeUnexpectedServerResponse CauseType = "UnexpectedServerResponse"
	CauseTypeFieldManagerConflict     CauseType = "FieldManagerConflict"
	CauseTypeResourceVersionTooLarge  CauseType = "ResourceVersionTooLarge"
)

type List struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items    []runtime.RawExtension `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type APIVersions struct {
	TypeMeta                   `json:",inline"`
	Versions                   []string                    `json:"versions" protobuf:"bytes,1,rep,name=versions"`
	ServerAddressByClientCIDRs []ServerAddressByClientCIDR `json:"serverAddressByClientCIDRs" protobuf:"bytes,2,rep,name=serverAddressByClientCIDRs"`
}

type APIGroupList struct {
	TypeMeta `json:",inline"`
	Groups   []APIGroup `json:"groups" protobuf:"bytes,1,rep,name=groups"`
}

type APIGroup struct {
	TypeMeta                   `json:",inline"`
	Name                       string                      `json:"name" protobuf:"bytes,1,opt,name=name"`
	Versions                   []GroupVersionForDiscovery  `json:"versions" protobuf:"bytes,2,rep,name=versions"`
	PreferredVersion           GroupVersionForDiscovery    `json:"preferredVersion,omitempty" protobuf:"bytes,3,opt,name=preferredVersion"`
	ServerAddressByClientCIDRs []ServerAddressByClientCIDR `json:"serverAddressByClientCIDRs,omitempty"`
}

type ServerAddressByClientCIDR struct {
	ClientCIDR    string `json:"clientCIDR" protobuf:"bytes,1,opt,name=clientCIDR"`
	ServerAddress string `json:"serverAddress" protobuf:"bytes,2,opt,name=serverAddress"`
}

type GroupVersionForDiscovery struct {
	GroupVersion string `json:"groupVersion" protobuf:"bytes,1,opt,name=groupVersion"`
	Version      string `json:"version" protobuf:"bytes,2,opt,name=version"`
}

type APIResource struct {
	Name               string   `json:"name" protobuf:"bytes,1,opt,name=name"`
	SingularName       string   `json:"singularName" protobuf:"bytes,6,opt,name=singularName"`
	Namespaced         bool     `json:"namespaced" protobuf:"varint,2,opt,name=namespaced"`
	Group              string   `json:"group,omitempty" protobuf:"bytes,8,opt,name=group"`
	Version            string   `json:"version,omitempty" protobuf:"bytes,9,opt,name=version"`
	Kind               string   `json:"kind" protobuf:"bytes,3,opt,name=kind"`
	Verbs              Verbs    `json:"verbs" protobuf:"bytes,4,opt,name=verbs"`
	ShortNames         []string `json:"shortNames,omitempty" protobuf:"bytes,5,rep,name=shortNames"`
	Categories         []string `json:"categories,omitempty" protobuf:"bytes,7,rep,name=categories"`
	StorageVersionHash string   `json:"storageVersionHash,omitempty" protobuf:"bytes,10,opt,name=storageVersionHash"`
}

type Verbs []string

func (vs Verbs) String() string {
	return fmt.Sprintf("%v", []string(vs))
}

type APIResourceList struct {
	TypeMeta     `json:",inline"`
	GroupVersion string        `json:"groupVersion" protobuf:"bytes,1,opt,name=groupVersion"`
	APIResources []APIResource `json:"resources" protobuf:"bytes,2,rep,name=resources"`
}

type RootPaths struct {
	Paths []string `json:"paths" protobuf:"bytes,1,rep,name=paths"`
}

func LabelSelectorQueryParam(version string) string {
	return "labelSelector"
}

func FieldSelectorQueryParam(version string) string {
	return "fieldSelector"
}

func (apiVersions APIVersions) String() string {
	return strings.Join(apiVersions.Versions, ",")
}

func (apiVersions APIVersions) GoString() string {
	return apiVersions.String()
}

type Patch struct{}

type LabelSelector struct {
	MatchLabels      map[string]string          `json:"matchLabels,omitempty" protobuf:"bytes,1,rep,name=matchLabels"`
	MatchExpressions []LabelSelectorRequirement `json:"matchExpressions,omitempty" protobuf:"bytes,2,rep,name=matchExpressions"`
}

type LabelSelectorRequirement struct {
	Key      string                `json:"key" patchStrategy:"merge" patchMergeKey:"key" protobuf:"bytes,1,opt,name=key"`
	Operator LabelSelectorOperator `json:"operator" protobuf:"bytes,2,opt,name=operator,casttype=LabelSelectorOperator"`
	Values   []string              `json:"values,omitempty" protobuf:"bytes,3,rep,name=values"`
}

type LabelSelectorOperator string

const (
	LabelSelectorOpIn           LabelSelectorOperator = "In"
	LabelSelectorOpNotIn        LabelSelectorOperator = "NotIn"
	LabelSelectorOpExists       LabelSelectorOperator = "Exists"
	LabelSelectorOpDoesNotExist LabelSelectorOperator = "DoesNotExist"
)

type ManagedFieldsEntry struct {
	Manager     string                     `json:"manager,omitempty" protobuf:"bytes,1,opt,name=manager"`
	Operation   ManagedFieldsOperationType `json:"operation,omitempty" protobuf:"bytes,2,opt,name=operation,casttype=ManagedFieldsOperationType"`
	APIVersion  string                     `json:"apiVersion,omitempty" protobuf:"bytes,3,opt,name=apiVersion"`
	Time        *Time                      `json:"time,omitempty" protobuf:"bytes,4,opt,name=time"`
	FieldsType  string                     `json:"fieldsType,omitempty" protobuf:"bytes,6,opt,name=fieldsType"`
	FieldsV1    *FieldsV1                  `json:"fieldsV1,omitempty" protobuf:"bytes,7,opt,name=fieldsV1"`
	Subresource string                     `json:"subresource,omitempty" protobuf:"bytes,8,opt,name=subresource"`
}

type ManagedFieldsOperationType string

const (
	ManagedFieldsOperationApply  ManagedFieldsOperationType = "Apply"
	ManagedFieldsOperationUpdate ManagedFieldsOperationType = "Update"
)

type FieldsV1 struct {
	Raw []byte `json:"-" protobuf:"bytes,1,opt,name=Raw"`
}

func (f FieldsV1) String() string {
	return string(f.Raw)
}

type Table struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata,omitempty"`

	ColumnDefinitions []TableColumnDefinition `json:"columnDefinitions"`
	Rows              []TableRow              `json:"rows"`
}

type TableColumnDefinition struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Format      string `json:"format"`
	Description string `json:"description"`
	Priority    int32  `json:"priority"`
}

type TableRow struct {
	Cells      []interface{}        `json:"cells"`
	Conditions []TableRowCondition  `json:"conditions,omitempty"`
	Object     runtime.RawExtension `json:"object,omitempty"`
}

type TableRowCondition struct {
	Type    RowConditionType `json:"type"`
	Status  ConditionStatus  `json:"status"`
	Reason  string           `json:"reason,omitempty"`
	Message string           `json:"message,omitempty"`
}

type RowConditionType string

const (
	RowCompleted RowConditionType = "Completed"
)

type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type IncludeObjectPolicy string

const (
	IncludeNone     IncludeObjectPolicy = "None"
	IncludeMetadata IncludeObjectPolicy = "Metadata"
	IncludeObject   IncludeObjectPolicy = "Object"
)

type TableOptions struct {
	TypeMeta      `json:",inline"`
	NoHeaders     bool                `json:"-"`
	IncludeObject IncludeObjectPolicy `json:"includeObject,omitempty" protobuf:"bytes,1,opt,name=includeObject,casttype=IncludeObjectPolicy"`
}

type PartialObjectMetadata struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

type PartialObjectMetadataList struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []PartialObjectMetadata `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type Condition struct {
	Type               string          `json:"type" protobuf:"bytes,1,opt,name=type"`
	Status             ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status"`
	ObservedGeneration int64           `json:"observedGeneration,omitempty" protobuf:"varint,3,opt,name=observedGeneration"`
	LastTransitionTime Time            `json:"lastTransitionTime" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	Reason             string          `json:"reason" protobuf:"bytes,5,opt,name=reason"`
	Message            string          `json:"message" protobuf:"bytes,6,opt,name=message"`
}
