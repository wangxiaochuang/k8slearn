package transport

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	utilnet "k8s.io/apimachinery/pkg/util/net"

	"k8s.io/apimachinery/pkg/util/wait"
)

type tlsTransportCache struct {
	mu         sync.Mutex
	transports map[tlsCacheKey]*http.Transport
}

const idleConnsPerHost = 25

var tlsCache = &tlsTransportCache{transports: make(map[tlsCacheKey]*http.Transport)}

type tlsCacheKey struct {
	insecure           bool
	caData             string
	certData           string
	keyData            string `datapolicy:"security-key"`
	certFile           string
	keyFile            string
	serverName         string
	nextProtos         string
	disableCompression bool
}

func (t tlsCacheKey) String() string {
	keyText := "<none>"
	if len(t.keyData) > 0 {
		keyText = "<redacted>"
	}
	return fmt.Sprintf("insecure:%v, caData:%#v, certData:%#v, keyData:%s, serverName:%s, disableCompression:%t", t.insecure, t.caData, t.certData, keyText, t.serverName, t.disableCompression)
}

func (c *tlsTransportCache) get(config *Config) (http.RoundTripper, error) {
	key, canCache, err := tlsConfigKey(config)
	if err != nil {
		return nil, err
	}

	if canCache {
		c.mu.Lock()
		defer c.mu.Unlock()

		if t, ok := c.transports[key]; ok {
			return t, nil
		}
	}

	tlsConfig, err := TLSConfigFor(config)
	if err != nil {
		return nil, err
	}

	if tlsConfig == nil && config.Dial == nil && config.Proxy == nil {
		return http.DefaultTransport, nil
	}

	dial := config.Dial
	if dial == nil {
		dial = (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext
	}

	if config.TLS.ReloadTLSFiles {
		dynamicCertDialer := certRotatingDialer(tlsConfig.GetClientCertificate, dial)
		tlsConfig.GetClientCertificate = dynamicCertDialer.GetClientCertificate
		dial = dynamicCertDialer.connDialer.DialContext
		go dynamicCertDialer.Run(wait.NeverStop)
	}

	proxy := http.ProxyFromEnvironment
	if config.Proxy != nil {
		proxy = config.Proxy
	}

	transport := utilnet.SetTransportDefaults(&http.Transport{
		Proxy:               proxy,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     tlsConfig,
		MaxIdleConnsPerHost: idleConnsPerHost,
		DialContext:         dial,
		DisableCompression:  config.DisableCompression,
	})

	if canCache {
		// Cache a single transport for these options
		c.transports[key] = transport
	}

	return transport, nil
}

func tlsConfigKey(c *Config) (tlsCacheKey, bool, error) {
	if err := loadTLSFiles(c); err != nil {
		return tlsCacheKey{}, false, err
	}

	if c.TLS.GetCert != nil || c.Dial != nil || c.Proxy != nil {
		// cannot determine equality for functions
		return tlsCacheKey{}, false, nil
	}

	k := tlsCacheKey{
		insecure:           c.TLS.Insecure,
		caData:             string(c.TLS.CAData),
		serverName:         c.TLS.ServerName,
		nextProtos:         strings.Join(c.TLS.NextProtos, ","),
		disableCompression: c.DisableCompression,
	}

	if c.TLS.ReloadTLSFiles {
		k.certFile = c.TLS.CertFile
		k.keyFile = c.TLS.KeyFile
	} else {
		k.certData = string(c.TLS.CertData)
		k.keyData = string(c.TLS.KeyData)
	}

	return k, true, nil
}
