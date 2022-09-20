package rest

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/transport"
	"k8s.io/client-go/util/flowcontrol"
)

const (
	DefaultQPS   float32 = 5.0
	DefaultBurst int     = 10
)

var ErrNotInCluster = errors.New("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")

type Config struct {
	Host    string
	APIPath string
	ContentConfig

	Username string
	Password string `datapolicy:"password"`

	BearerToken         string `datapolicy:"token"`
	BearerTokenFile     string
	Impersonate         ImpersonationConfig
	AuthProvider        *clientcmdapi.AuthProviderConfig
	AuthConfigPersister AuthProviderConfigPersister
	ExecProvider        *clientcmdapi.ExecConfig
	TLSClientConfig
	UserAgent          string
	DisableCompression bool
	Transport          http.RoundTripper
	WrapTransport      transport.WrapperFunc
	QPS                float32
	Burst              int
	RateLimiter        flowcontrol.RateLimiter
	WarningHandler     WarningHandler
	Timeout            time.Duration
	Dial               func(ctx context.Context, network, address string) (net.Conn, error)
	Proxy              func(*http.Request) (*url.URL, error)
}

var _ fmt.Stringer = new(Config)
var _ fmt.GoStringer = new(Config)

type sanitizedConfig *Config

type sanitizedAuthConfigPersister struct{ AuthProviderConfigPersister }

func (sanitizedAuthConfigPersister) GoString() string {
	return "rest.AuthProviderConfigPersister(--- REDACTED ---)"
}
func (sanitizedAuthConfigPersister) String() string {
	return "rest.AuthProviderConfigPersister(--- REDACTED ---)"
}

type sanitizedObject struct{ runtime.Object }

func (sanitizedObject) GoString() string {
	return "runtime.Object(--- REDACTED ---)"
}
func (sanitizedObject) String() string {
	return "runtime.Object(--- REDACTED ---)"
}

func (c *Config) GoString() string {
	return c.String()
}

func (c *Config) String() string {
	if c == nil {
		return "<nil>"
	}
	cc := sanitizedConfig(CopyConfig(c))
	// Explicitly mark non-empty credential fields as redacted.
	if cc.Password != "" {
		cc.Password = "--- REDACTED ---"
	}
	if cc.BearerToken != "" {
		cc.BearerToken = "--- REDACTED ---"
	}
	if cc.AuthConfigPersister != nil {
		cc.AuthConfigPersister = sanitizedAuthConfigPersister{cc.AuthConfigPersister}
	}
	if cc.ExecProvider != nil && cc.ExecProvider.Config != nil {
		cc.ExecProvider.Config = sanitizedObject{Object: cc.ExecProvider.Config}
	}
	return fmt.Sprintf("%#v", cc)
}

type ImpersonationConfig struct {
	UserName string
	UID      string
	Groups   []string
	Extra    map[string][]string
}

type TLSClientConfig struct {
	Insecure   bool
	ServerName string
	CertFile   string
	KeyFile    string
	CAFile     string
	CertData   []byte
	KeyData    []byte `datapolicy:"security-key"`
	CAData     []byte
	NextProtos []string
}

var _ fmt.Stringer = TLSClientConfig{}
var _ fmt.GoStringer = TLSClientConfig{}

type sanitizedTLSClientConfig TLSClientConfig

func (c TLSClientConfig) GoString() string {
	return c.String()
}

func (c TLSClientConfig) String() string {
	cc := sanitizedTLSClientConfig{
		Insecure:   c.Insecure,
		ServerName: c.ServerName,
		CertFile:   c.CertFile,
		KeyFile:    c.KeyFile,
		CAFile:     c.CAFile,
		CertData:   c.CertData,
		KeyData:    c.KeyData,
		CAData:     c.CAData,
		NextProtos: c.NextProtos,
	}
	if len(cc.CertData) != 0 {
		cc.CertData = []byte("--- TRUNCATED ---")
	}
	if len(cc.KeyData) != 0 {
		cc.KeyData = []byte("--- REDACTED ---")
	}
	return fmt.Sprintf("%#v", cc)
}

type ContentConfig struct {
	AcceptContentTypes   string
	ContentType          string
	GroupVersion         *schema.GroupVersion
	NegotiatedSerializer runtime.NegotiatedSerializer
}

func CopyConfig(config *Config) *Config {
	c := &Config{
		Host:            config.Host,
		APIPath:         config.APIPath,
		ContentConfig:   config.ContentConfig,
		Username:        config.Username,
		Password:        config.Password,
		BearerToken:     config.BearerToken,
		BearerTokenFile: config.BearerTokenFile,
		Impersonate: ImpersonationConfig{
			UserName: config.Impersonate.UserName,
			UID:      config.Impersonate.UID,
			Groups:   config.Impersonate.Groups,
			Extra:    config.Impersonate.Extra,
		},
		AuthProvider:        config.AuthProvider,
		AuthConfigPersister: config.AuthConfigPersister,
		ExecProvider:        config.ExecProvider,
		TLSClientConfig: TLSClientConfig{
			Insecure:   config.TLSClientConfig.Insecure,
			ServerName: config.TLSClientConfig.ServerName,
			CertFile:   config.TLSClientConfig.CertFile,
			KeyFile:    config.TLSClientConfig.KeyFile,
			CAFile:     config.TLSClientConfig.CAFile,
			CertData:   config.TLSClientConfig.CertData,
			KeyData:    config.TLSClientConfig.KeyData,
			CAData:     config.TLSClientConfig.CAData,
			NextProtos: config.TLSClientConfig.NextProtos,
		},
		UserAgent:          config.UserAgent,
		DisableCompression: config.DisableCompression,
		Transport:          config.Transport,
		WrapTransport:      config.WrapTransport,
		QPS:                config.QPS,
		Burst:              config.Burst,
		RateLimiter:        config.RateLimiter,
		WarningHandler:     config.WarningHandler,
		Timeout:            config.Timeout,
		Dial:               config.Dial,
		Proxy:              config.Proxy,
	}
	if config.ExecProvider != nil && config.ExecProvider.Config != nil {
		c.ExecProvider.Config = config.ExecProvider.Config.DeepCopyObject()
	}
	return c
}
