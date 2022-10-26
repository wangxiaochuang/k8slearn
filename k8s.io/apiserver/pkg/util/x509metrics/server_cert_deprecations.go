package x509metrics

import (
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apiserver/pkg/audit"
	"k8s.io/component-base/metrics"
	"k8s.io/klog/v2"
)

var _ utilnet.RoundTripperWrapper = &x509DeprecatedCertificateMetricsRTWrapper{}

type x509DeprecatedCertificateMetricsRTWrapper struct {
	rt http.RoundTripper

	checkers []deprecatedCertificateAttributeChecker
}

type deprecatedCertificateAttributeChecker interface {
	CheckRoundTripError(err error) bool
	CheckPeerCertificates(certs []*x509.Certificate) bool
	IncreaseMetricsCounter(req *http.Request)
}

type counterRaiser struct {
	counter *metrics.Counter
	id      string
	reason  string
}

func (c *counterRaiser) IncreaseMetricsCounter(req *http.Request) {
	if req != nil && req.URL != nil {
		if hostname := req.URL.Hostname(); len(hostname) > 0 {
			prefix := fmt.Sprintf("%s.invalid-cert.kubernetes.io", c.id)
			klog.Infof("%s: invalid certificate detected connecting to %q: %s", prefix, hostname, c.reason)
			audit.AddAuditAnnotation(req.Context(), prefix+"/"+hostname, c.reason)
		}
	}
	c.counter.Inc()
}

func NewDeprecatedCertificateRoundTripperWrapperConstructor(missingSAN, sha1 *metrics.Counter) func(rt http.RoundTripper) http.RoundTripper {
	return func(rt http.RoundTripper) http.RoundTripper {
		return &x509DeprecatedCertificateMetricsRTWrapper{
			rt: rt,
			checkers: []deprecatedCertificateAttributeChecker{
				NewSANDeprecatedChecker(missingSAN),
				NewSHA1SignatureDeprecatedChecker(sha1),
			},
		}
	}
}

func (w *x509DeprecatedCertificateMetricsRTWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	panic("not implemented")
}

func (w *x509DeprecatedCertificateMetricsRTWrapper) WrappedRoundTripper() http.RoundTripper {
	return w.rt
}

var _ deprecatedCertificateAttributeChecker = &missingSANChecker{}

type missingSANChecker struct {
	counterRaiser
}

func NewSANDeprecatedChecker(counter *metrics.Counter) *missingSANChecker {
	return &missingSANChecker{
		counterRaiser: counterRaiser{
			counter: counter,
			id:      "missing-san",
			reason:  "relies on a legacy Common Name field instead of the SAN extension for subject validation",
		},
	}
}

func (c *missingSANChecker) CheckRoundTripError(err error) bool {
	if err != nil && errors.As(err, &x509.HostnameError{}) && strings.Contains(err.Error(), "x509: certificate relies on legacy Common Name field") {
		// increase the count of registered failures due to Go 1.15 x509 cert Common Name deprecation
		return true
	}

	return false
}

func (c *missingSANChecker) CheckPeerCertificates(peerCertificates []*x509.Certificate) bool {
	if len(peerCertificates) > 0 {
		if serverCert := peerCertificates[0]; !hasSAN(serverCert) {
			return true
		}
	}

	return false
}

func hasSAN(c *x509.Certificate) bool {
	sanOID := []int{2, 5, 29, 17}

	for _, e := range c.Extensions {
		if e.Id.Equal(sanOID) {
			return true
		}
	}
	return false
}

type sha1SignatureChecker struct {
	*counterRaiser
}

func NewSHA1SignatureDeprecatedChecker(counter *metrics.Counter) *sha1SignatureChecker {
	return &sha1SignatureChecker{
		counterRaiser: &counterRaiser{
			counter: counter,
			id:      "insecure-sha1",
			reason:  "uses an insecure SHA-1 signature",
		},
	}
}

func (c *sha1SignatureChecker) CheckRoundTripError(err error) bool {
	var unknownAuthorityError x509.UnknownAuthorityError
	if err == nil {
		return false
	}
	if !errors.As(err, &unknownAuthorityError) {
		return false
	}

	errMsg := err.Error()
	if strIdx := strings.Index(errMsg, "x509: cannot verify signature: insecure algorithm"); strIdx != -1 && strings.Contains(errMsg[strIdx:], "SHA1") {
		// increase the count of registered failures due to Go 1.18 x509 sha1 signature deprecation
		return true
	}

	return false
}

func (c *sha1SignatureChecker) CheckPeerCertificates(peerCertificates []*x509.Certificate) bool {
	// check all received non-self-signed certificates for deprecated signing algorithms
	for _, cert := range peerCertificates {
		if cert.SignatureAlgorithm == x509.SHA1WithRSA || cert.SignatureAlgorithm == x509.ECDSAWithSHA1 {
			// the SHA-1 deprecation does not involve self-signed root certificates
			if !reflect.DeepEqual(cert.Issuer, cert.Subject) {
				return true
			}
		}
	}

	return false
}
