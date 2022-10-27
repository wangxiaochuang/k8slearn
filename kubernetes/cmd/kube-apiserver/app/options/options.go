package options

import (
	"net"
	"strings"
	"time"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/kubernetes/pkg/controlplane/reconcilers"
	_ "k8s.io/kubernetes/pkg/features"
	kubeoptions "k8s.io/kubernetes/pkg/kubeapiserver/options"
	"k8s.io/kubernetes/pkg/serviceaccount"
)

type ServerRunOptions struct {
	GenericServerRunOptions *genericoptions.ServerRunOptions
	Etcd                    *genericoptions.EtcdOptions
	SecureServing           *genericoptions.SecureServingOptionsWithLoopback
	Audit                   *genericoptions.AuditOptions
	Features                *genericoptions.FeatureOptions
	Admission               *kubeoptions.AdmissionOptions
	Authentication          *kubeoptions.BuiltInAuthenticationOptions
	Authorization           *kubeoptions.BuiltInAuthorizationOptions
	CloudProvider           *kubeoptions.CloudProviderOptions
	APIEnablement           *genericoptions.APIEnablementOptions
	EgressSelector          *genericoptions.EgressSelectorOptions

	Logs   *logs.Options
	Traces *genericoptions.TracingOptions

	AllowPrivileged   bool
	EnableLogsHandler bool
	EventTTL          time.Duration
	// KubeletConfig             kubeletclient.KubeletClientConfig
	KubernetesServiceNodePort int
	MaxConnectionBytesPerSec  int64

	ServiceClusterIPRanges         string
	PrimaryServiceClusterIPRange   net.IPNet
	SecondaryServiceClusterIPRange net.IPNet
	// APIServerServiceIP is the first valid IP from PrimaryServiceClusterIPRange
	APIServerServiceIP net.IP

	ServiceNodePortRange utilnet.PortRange

	ProxyClientCertFile string
	ProxyClientKeyFile  string

	EnableAggregatorRouting bool

	MasterCount            int
	EndpointReconcilerType string

	IdentityLeaseDurationSeconds      int
	IdentityLeaseRenewIntervalSeconds int

	ServiceAccountSigningKeyFile     string
	ServiceAccountIssuer             serviceaccount.TokenGenerator
	ServiceAccountTokenMaxExpiration time.Duration

	ShowHiddenMetricsForVersion string
}

func NewServerRunOptions() *ServerRunOptions {
	s := ServerRunOptions{
		GenericServerRunOptions:           genericoptions.NewServerRunOptions(),
		Etcd:                              genericoptions.NewEtcdOptions(storagebackend.NewDefaultConfig(kubeoptions.DefaultEtcdPathPrefix, nil)),
		SecureServing:                     kubeoptions.NewSecureServingOptions(),
		Audit:                             genericoptions.NewAuditOptions(),
		Features:                          genericoptions.NewFeatureOptions(),
		Admission:                         kubeoptions.NewAdmissionOptions(),
		Authentication:                    kubeoptions.NewBuiltInAuthenticationOptions().WithAll(),
		Authorization:                     kubeoptions.NewBuiltInAuthorizationOptions(),
		CloudProvider:                     kubeoptions.NewCloudProviderOptions(),
		APIEnablement:                     genericoptions.NewAPIEnablementOptions(),
		EgressSelector:                    genericoptions.NewEgressSelectorOptions(),
		Logs:                              logs.NewOptions(),
		Traces:                            genericoptions.NewTracingOptions(),
		EventTTL:                          1 * time.Hour,
		MasterCount:                       1,
		EndpointReconcilerType:            string(reconcilers.LeaseEndpointReconcilerType),
		IdentityLeaseDurationSeconds:      3600,
		IdentityLeaseRenewIntervalSeconds: 10,

		ServiceNodePortRange: kubeoptions.DefaultServiceNodePortRange,
	}

	s.Etcd.DefaultStorageMediaType = "application/vnd.kubernetes.protobuf"
	return &s
}

func (s *ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {
	s.GenericServerRunOptions.AddUniversalFlags(fss.FlagSet("generic"))
	s.Etcd.AddFlags(fss.FlagSet("etcd"))
	s.SecureServing.AddFlags(fss.FlagSet("secure serving"))
	s.Audit.AddFlags(fss.FlagSet("auditing"))
	s.Features.AddFlags(fss.FlagSet("features"))
	s.Authentication.AddFlags(fss.FlagSet("authentication"))
	s.Authorization.AddFlags(fss.FlagSet("authorization"))
	s.CloudProvider.AddFlags(fss.FlagSet("cloud provider"))
	s.APIEnablement.AddFlags(fss.FlagSet("API enablement"))
	s.EgressSelector.AddFlags(fss.FlagSet("egress selector"))
	s.Admission.AddFlags(fss.FlagSet("admission"))

	// --log-flush-frequency --log-json-info-buffer-size --log-json-split-stream --logging-format -v --vmodule
	s.Logs.AddFlags(fss.FlagSet("logs"))
	s.Traces.AddFlags(fss.FlagSet("traces"))

	fs := fss.FlagSet("misc")
	fs.DurationVar(&s.EventTTL, "event-ttl", s.EventTTL,
		"Amount of time to retain events.")

	fs.BoolVar(&s.AllowPrivileged, "allow-privileged", s.AllowPrivileged,
		"If true, allow privileged containers. [default=false]")

	fs.BoolVar(&s.EnableLogsHandler, "enable-logs-handler", s.EnableLogsHandler,
		"If true, install a /logs handler for the apiserver logs.")
	fs.MarkDeprecated("enable-logs-handler", "This flag will be removed in v1.19")

	fs.Int64Var(&s.MaxConnectionBytesPerSec, "max-connection-bytes-per-sec", s.MaxConnectionBytesPerSec, ""+
		"If non-zero, throttle each user connection to this number of bytes/sec. "+
		"Currently only applies to long-running requests.")

	fs.IntVar(&s.MasterCount, "apiserver-count", s.MasterCount,
		"The number of apiservers running in the cluster, must be a positive number. (In use when --endpoint-reconciler-type=master-count is enabled.)")

	fs.MarkDeprecated("apiserver-count", "apiserver-count is deprecated and will be removed in a future version.")

	fs.StringVar(&s.EndpointReconcilerType, "endpoint-reconciler-type", string(s.EndpointReconcilerType),
		"Use an endpoint reconciler ("+strings.Join(reconcilers.AllTypes.Names(), ", ")+") master-count is deprecated, and will be removed in a future version.")

	fs.IntVar(&s.IdentityLeaseDurationSeconds, "identity-lease-duration-seconds", s.IdentityLeaseDurationSeconds,
		"The duration of kube-apiserver lease in seconds, must be a positive number. (In use when the APIServerIdentity feature gate is enabled.)")

	fs.IntVar(&s.IdentityLeaseRenewIntervalSeconds, "identity-lease-renew-interval-seconds", s.IdentityLeaseRenewIntervalSeconds,
		"The interval of kube-apiserver renewing its lease in seconds, must be a positive number. (In use when the APIServerIdentity feature gate is enabled.)")

	fs.IntVar(&s.KubernetesServiceNodePort, "kubernetes-service-node-port", s.KubernetesServiceNodePort, ""+
		"If non-zero, the Kubernetes master service (which apiserver creates/maintains) will be "+
		"of type NodePort, using this as the value of the port. If zero, the Kubernetes master "+
		"service will be of type ClusterIP.")

	fs.StringVar(&s.ServiceClusterIPRanges, "service-cluster-ip-range", s.ServiceClusterIPRanges, ""+
		"A CIDR notation IP range from which to assign service cluster IPs. This must not "+
		"overlap with any IP ranges assigned to nodes or pods. Max of two dual-stack CIDRs is allowed.")

	fs.Var(&s.ServiceNodePortRange, "service-node-port-range", ""+
		"A port range to reserve for services with NodePort visibility. "+
		"Example: '30000-32767'. Inclusive at both ends of the range.")

	fs.StringVar(&s.ProxyClientCertFile, "proxy-client-cert-file", s.ProxyClientCertFile, ""+
		"Client certificate used to prove the identity of the aggregator or kube-apiserver "+
		"when it must call out during a request. This includes proxying requests to a user "+
		"api-server and calling out to webhook admission plugins. It is expected that this "+
		"cert includes a signature from the CA in the --requestheader-client-ca-file flag. "+
		"That CA is published in the 'extension-apiserver-authentication' configmap in "+
		"the kube-system namespace. Components receiving calls from kube-aggregator should "+
		"use that CA to perform their half of the mutual TLS verification.")
	fs.StringVar(&s.ProxyClientKeyFile, "proxy-client-key-file", s.ProxyClientKeyFile, ""+
		"Private key for the client certificate used to prove the identity of the aggregator or kube-apiserver "+
		"when it must call out during a request. This includes proxying requests to a user "+
		"api-server and calling out to webhook admission plugins.")

	fs.BoolVar(&s.EnableAggregatorRouting, "enable-aggregator-routing", s.EnableAggregatorRouting,
		"Turns on aggregator routing requests to endpoints IP rather than cluster IP.")

	fs.StringVar(&s.ServiceAccountSigningKeyFile, "service-account-signing-key-file", s.ServiceAccountSigningKeyFile, ""+
		"Path to the file that contains the current private key of the service account token issuer. The issuer will sign issued ID tokens with this private key.")

	return fss
}
