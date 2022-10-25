package transport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/klog/v2"
)

func New(config *Config) (http.RoundTripper, error) {
	if config.Transport != nil && (config.HasCA() || config.HasCertAuth() || config.HasCertCallback() || config.TLS.Insecure) {
		return nil, fmt.Errorf("using a custom transport with TLS certificate options or the insecure flag is not allowed")
	}

	var (
		rt  http.RoundTripper
		err error
	)

	if config.Transport != nil {
		rt = config.Transport
	} else {
		rt, err = tlsCache.get(config)
		if err != nil {
			return nil, err
		}
	}
	return HTTPWrappersForConfig(config, rt)
}

// p61
func TLSConfigFor(c *Config) (*tls.Config, error) {
	if !(c.HasCA() || c.HasCertAuth() || c.HasCertCallback() || c.TLS.Insecure || len(c.TLS.ServerName) > 0 || len(c.TLS.NextProtos) > 0) {
		return nil, nil
	}
	if c.HasCA() && c.TLS.Insecure {
		return nil, fmt.Errorf("specifying a root certificates file with the insecure flag is not allowed")
	}
	if err := loadTLSFiles(c); err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		// Can't use SSLv3 because of POODLE and BEAST
		// Can't use TLSv1.0 because of POODLE and BEAST using CBC cipher
		// Can't use TLSv1.1 because of RC4 cipher usage
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: c.TLS.Insecure,
		ServerName:         c.TLS.ServerName,
		NextProtos:         c.TLS.NextProtos,
	}

	if c.HasCA() {
		rootCAs, err := rootCertPool(c.TLS.CAData)
		if err != nil {
			return nil, fmt.Errorf("unable to load root certificates: %w", err)
		}
		tlsConfig.RootCAs = rootCAs
	}

	var staticCert *tls.Certificate

	if c.HasCertAuth() && !c.TLS.ReloadTLSFiles {
		// If key/cert were provided, verify them before setting up
		// tlsConfig.GetClientCertificate.
		cert, err := tls.X509KeyPair(c.TLS.CertData, c.TLS.KeyData)
		if err != nil {
			return nil, err
		}
		staticCert = &cert
	}

	var dynamicCertLoader func() (*tls.Certificate, error)
	if c.TLS.ReloadTLSFiles {
		dynamicCertLoader = cachingCertificateLoader(c.TLS.CertFile, c.TLS.KeyFile)
	}

	if c.HasCertAuth() || c.HasCertCallback() {
		tlsConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			if staticCert != nil {
				return staticCert, nil
			}

			if dynamicCertLoader != nil {
				return dynamicCertLoader()
			}

			if c.HasCertCallback() {
				cert, err := c.TLS.GetCert()
				if err != nil {
					return nil, err
				}
				// GetCert may return empty value, meaning no cert.
				if cert != nil {
					return cert, nil
				}
			}

			return &tls.Certificate{}, nil
		}
	}

	return tlsConfig, nil
}

// p142
func loadTLSFiles(c *Config) error {
	var err error
	c.TLS.CAData, err = dataFromSliceOrFile(c.TLS.CAData, c.TLS.CAFile)
	if err != nil {
		return err
	}

	if len(c.TLS.CertFile) > 0 && len(c.TLS.CertData) == 0 && len(c.TLS.KeyFile) > 0 && len(c.TLS.KeyData) == 0 {
		c.TLS.ReloadTLSFiles = true
	}

	c.TLS.CertData, err = dataFromSliceOrFile(c.TLS.CertData, c.TLS.CertFile)
	if err != nil {
		return err
	}

	c.TLS.KeyData, err = dataFromSliceOrFile(c.TLS.KeyData, c.TLS.KeyFile)
	if err != nil {
		return err
	}
	return nil
}

func dataFromSliceOrFile(data []byte, file string) ([]byte, error) {
	if len(data) > 0 {
		return data, nil
	}
	if len(file) > 0 {
		fileData, err := ioutil.ReadFile(file)
		if err != nil {
			return []byte{}, err
		}
		return fileData, nil
	}
	return nil, nil
}

func rootCertPool(caData []byte) (*x509.CertPool, error) {
	if len(caData) == 0 {
		return nil, nil
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caData); !ok {
		return nil, createErrorParsingCAData(caData)
	}
	return certPool, nil
}

func createErrorParsingCAData(pemCerts []byte) error {
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			return fmt.Errorf("unable to parse bytes as PEM block")
		}

		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		if _, err := x509.ParseCertificate(block.Bytes); err != nil {
			return fmt.Errorf("failed to parse certificate: %w", err)
		}
	}
	return fmt.Errorf("no valid certificate authority data seen")
}

// p224
type WrapperFunc func(rt http.RoundTripper) http.RoundTripper

func Wrappers(fns ...WrapperFunc) WrapperFunc {
	if len(fns) == 0 {
		return nil
	}
	if len(fns) == 2 && fns[0] == nil {
		return fns[1]
	}
	return func(rt http.RoundTripper) http.RoundTripper {
		base := rt
		for _, fn := range fns {
			if fn != nil {
				base = fn(base)
			}
		}
		return base
	}
}

func ContextCanceller(ctx context.Context, err error) WrapperFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		return &contextCanceller{
			ctx: ctx,
			rt:  rt,
			err: err,
		}
	}
}

type contextCanceller struct {
	ctx context.Context
	rt  http.RoundTripper
	err error
}

func (b *contextCanceller) RoundTrip(req *http.Request) (*http.Response, error) {
	select {
	case <-b.ctx.Done():
		return nil, b.err
	default:
		return b.rt.RoundTrip(req)
	}
}

// p278
func tryCancelRequest(rt http.RoundTripper, req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}
	switch rt := rt.(type) {
	case canceler:
		rt.CancelRequest(req)
	case utilnet.RoundTripperWrapper:
		tryCancelRequest(rt.WrappedRoundTripper(), req)
	default:
		klog.Warningf("Unable to cancel request for %T", rt)
	}
}

type certificateCacheEntry struct {
	cert  *tls.Certificate
	err   error
	birth time.Time
}

// isStale returns true when this cache entry is too old to be usable
func (c *certificateCacheEntry) isStale() bool {
	return time.Since(c.birth) > time.Second
}

func newCertificateCacheEntry(certFile, keyFile string) certificateCacheEntry {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	return certificateCacheEntry{cert: &cert, err: err, birth: time.Now()}
}

// p310 缓存一下，防止频繁得读取文件系统，最多每秒读取一次
func cachingCertificateLoader(certFile, keyFile string) func() (*tls.Certificate, error) {
	current := newCertificateCacheEntry(certFile, keyFile)
	var currentMtx sync.RWMutex

	return func() (*tls.Certificate, error) {
		currentMtx.RLock()
		if current.isStale() {
			currentMtx.RUnlock()

			currentMtx.Lock()
			defer currentMtx.Unlock()

			if current.isStale() {
				current = newCertificateCacheEntry(certFile, keyFile)
			}
		} else {
			defer currentMtx.RUnlock()
		}

		return current.cert, current.err
	}
}
