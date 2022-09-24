package server

import (
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	utilwaitgroup "k8s.io/apimachinery/pkg/util/waitgroup"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/audit"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/endpoints/discovery"
	apiopenapi "k8s.io/apiserver/pkg/endpoints/openapi"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/features"
	genericregistry "k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
	genericfilters "k8s.io/apiserver/pkg/server/filters"
	"k8s.io/apiserver/pkg/server/healthz"
	serverstore "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storageversion"
	"k8s.io/apiserver/pkg/util/feature"
	utilflowcontrol "k8s.io/apiserver/pkg/util/flowcontrol"
	flowcontrolrequest "k8s.io/apiserver/pkg/util/flowcontrol/request"
	openapicommon "k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/validation/spec"

	"go.opentelemetry.io/otel/trace"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/server/egressselector"
	"k8s.io/client-go/informers"
	restclient "k8s.io/client-go/rest"
)

const (
	DefaultLegacyAPIPrefix = "/api"
	APIGroupPrefix         = "/apis"
)

type Config struct {
	SecureServing             *SecureServingInfo
	Authentication            AuthenticationInfo
	Authorization             AuthorizationInfo
	LoopbackClientConfig      *restclient.Config
	EgressSelector            *egressselector.EgressSelector
	RuleResolver              authorizer.RuleResolver
	AdmissionControl          admission.Interface
	CorsAllowedOriginList     []string
	HSTSDirectives            []string
	FlowControl               utilflowcontrol.Interface
	EnableIndex               bool
	EnableProfiling           bool
	EnableDiscovery           bool
	EnableContentionProfiling bool
	EnableMetrics             bool

	DisabledPostStartHooks   sets.String
	PostStartHooks           map[string]PostStartHookConfigEntry
	Version                  *version.Info
	AuditBackend             audit.Backend
	AuditPolicyRuleEvaluator audit.PolicyRuleEvaluator
	ExternalAddress          string

	TracerProvider              *trace.TracerProvider
	BuildHandlerChainFunc       func(apiHandler http.Handler, c *Config) (secure http.Handler)
	HandlerChainWaitGroup       *utilwaitgroup.SafeWaitGroup
	DiscoveryAddresses          discovery.Addresses
	HealthzChecks               []healthz.HealthChecker
	LivezChecks                 []healthz.HealthChecker
	ReadyzChecks                []healthz.HealthChecker
	LegacyAPIGroupPrefixes      sets.String
	RequestInfoResolver         apirequest.RequestInfoResolver
	Serializer                  runtime.NegotiatedSerializer
	OpenAPIConfig               *openapicommon.Config
	OpenAPIV3Config             *openapicommon.Config
	SkipOpenAPIInstallation     bool
	RESTOptionsGetter           genericregistry.RESTOptionsGetter
	RequestTimeout              time.Duration
	MinRequestTimeout           int
	LivezGracePeriod            time.Duration
	ShutdownDelayDuration       time.Duration
	JSONPatchMaxCopyBytes       int64
	MaxRequestBodyBytes         int64
	MaxRequestsInFlight         int
	MaxMutatingRequestsInFlight int
	LongRunningFunc             apirequest.LongRunningRequestCheck
	GoawayChance                float64
	MergedResourceConfig        *serverstore.ResourceConfig
	lifecycleSignals            lifecycleSignals
	StorageObjectCountTracker   flowcontrolrequest.StorageObjectCountTracker
	ShutdownSendRetryAfter      bool
	PublicAddress               net.IP
	EquivalentResourceRegistry  runtime.EquivalentResourceRegistry
	APIServerID                 string
	StorageVersionManager       storageversion.Manager
}

type RecommendedConfig struct {
	Config
	SharedInformerFactory informers.SharedInformerFactory
	ClientConfig          *restclient.Config
}

type SecureServingInfo struct {
	Listener                     net.Listener
	Cert                         dynamiccertificates.CertKeyContentProvider
	SNICerts                     []dynamiccertificates.SNICertKeyContentProvider
	ClientCA                     dynamiccertificates.CAContentProvider
	MinTLSVersion                uint16
	CipherSuites                 []uint16
	HTTP2MaxStreamsPerConnection int
	DisableHTTP2                 bool
}

type AuthenticationInfo struct {
	APIAudiences  authenticator.Audiences
	Authenticator authenticator.Request
}

type AuthorizationInfo struct {
	Authorizer authorizer.Authorizer
}

// p323
func NewConfig(codecs serializer.CodecFactory) *Config {
	defaultHealthChecks := []healthz.HealthChecker{healthz.PingHealthz, healthz.LogHealthz}
	var id string
	if feature.DefaultFeatureGate.Enabled(features.APIServerIdentity) {
		id = "kube-apiserver-" + uuid.New().String()
	}
	lifecycleSignals := newLifecycleSignals()

	return &Config{
		Serializer:                  codecs,
		BuildHandlerChainFunc:       DefaultBuildHandlerChain,
		HandlerChainWaitGroup:       new(utilwaitgroup.SafeWaitGroup),
		LegacyAPIGroupPrefixes:      sets.NewString(DefaultLegacyAPIPrefix),
		DisabledPostStartHooks:      sets.NewString(),
		PostStartHooks:              map[string]PostStartHookConfigEntry{},
		HealthzChecks:               append([]healthz.HealthChecker{}, defaultHealthChecks...),
		ReadyzChecks:                append([]healthz.HealthChecker{}, defaultHealthChecks...),
		LivezChecks:                 append([]healthz.HealthChecker{}, defaultHealthChecks...),
		EnableIndex:                 true,
		EnableDiscovery:             true,
		EnableProfiling:             true,
		EnableMetrics:               true,
		MaxRequestsInFlight:         400,
		MaxMutatingRequestsInFlight: 200,
		RequestTimeout:              time.Duration(60) * time.Second,
		MinRequestTimeout:           1800,
		LivezGracePeriod:            time.Duration(0),
		ShutdownDelayDuration:       time.Duration(0),
		JSONPatchMaxCopyBytes:       int64(3 * 1024 * 1024),
		MaxRequestBodyBytes:         int64(3 * 1024 * 1024),

		LongRunningFunc:           genericfilters.BasicLongRunningRequestCheck(sets.NewString("watch"), sets.NewString()),
		lifecycleSignals:          lifecycleSignals,
		StorageObjectCountTracker: flowcontrolrequest.NewStorageObjectCountTracker(lifecycleSignals.ShutdownInitiated.Signaled()),

		APIServerID:           id,
		StorageVersionManager: storageversion.NewDefaultManager(),
	}
}

func NewRecommendedConfig(codecs serializer.CodecFactory) *RecommendedConfig {
	return &RecommendedConfig{
		Config: *NewConfig(codecs),
	}
}

// p388
func DefaultOpenAPIConfig(getDefinitions openapicommon.GetOpenAPIDefinitions, defNamer *apiopenapi.DefinitionNamer) *openapicommon.Config {
	return &openapicommon.Config{
		ProtocolList:   []string{"https"},
		IgnorePrefixes: []string{},
		Info: &spec.Info{
			InfoProps: spec.InfoProps{
				Title: "Generic API Server",
			},
		},
		DefaultResponse: &spec.Response{
			ResponseProps: spec.ResponseProps{
				Description: "Default Response.",
			},
		},
		GetOperationIDAndTags: apiopenapi.GetOperationIDAndTags,
		GetDefinitionName:     defNamer.GetDefinitionName,
		GetDefinitions:        getDefinitions,
	}
}

// p409
func DefaultOpenAPIV3Config(getDefinitions openapicommon.GetOpenAPIDefinitions, defNamer *apiopenapi.DefinitionNamer) *openapicommon.Config {
	defaultConfig := DefaultOpenAPIConfig(getDefinitions, defNamer)
	defaultConfig.Definitions = getDefinitions(func(name string) spec.Ref {
		defName, _ := defaultConfig.GetDefinitionName(name)
		return spec.MustCreateRef("#/components/schemas/" + openapicommon.EscapeJsonPointer(defName))
	})

	return defaultConfig
}

// p799
func DefaultBuildHandlerChain(apiHandler http.Handler, c *Config) http.Handler {
	panic("xxxxxx")
	return nil
}
