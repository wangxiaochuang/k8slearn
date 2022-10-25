package x509

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

var clientCertificateExpirationHistogram = metrics.NewHistogram(
	&metrics.HistogramOpts{
		Namespace: "apiserver",
		Subsystem: "client",
		Name:      "certificate_expiration_seconds",
		Help:      "Distribution of the remaining lifetime on the certificate used to authenticate a request.",
		Buckets: []float64{
			0,
			(30 * time.Minute).Seconds(),
			(1 * time.Hour).Seconds(),
			(2 * time.Hour).Seconds(),
			(6 * time.Hour).Seconds(),
			(12 * time.Hour).Seconds(),
			(24 * time.Hour).Seconds(),
			(2 * 24 * time.Hour).Seconds(),
			(4 * 24 * time.Hour).Seconds(),
			(7 * 24 * time.Hour).Seconds(),
			(30 * 24 * time.Hour).Seconds(),
			(3 * 30 * 24 * time.Hour).Seconds(),
			(6 * 30 * 24 * time.Hour).Seconds(),
			(12 * 30 * 24 * time.Hour).Seconds(),
		},
		StabilityLevel: metrics.ALPHA,
	},
)

func init() {
	legacyregistry.MustRegister(clientCertificateExpirationHistogram)
}

type UserConversion interface {
	User(chain []*x509.Certificate) (*authenticator.Response, bool, error)
}

type UserConversionFunc func(chain []*x509.Certificate) (*authenticator.Response, bool, error)

func (f UserConversionFunc) User(chain []*x509.Certificate) (*authenticator.Response, bool, error) {
	return f(chain)
}

func columnSeparatedHex(d []byte) string {
	h := strings.ToUpper(hex.EncodeToString(d))
	var sb strings.Builder
	for i, r := range h {
		sb.WriteRune(r)
		if i%2 == 1 && i != len(h)-1 {
			sb.WriteRune(':')
		}
	}
	return sb.String()
}

func certificateIdentifier(c *x509.Certificate) string {
	return fmt.Sprintf(
		"SN=%d, SKID=%s, AKID=%s",
		c.SerialNumber,
		columnSeparatedHex(c.SubjectKeyId),
		columnSeparatedHex(c.AuthorityKeyId),
	)
}

type VerifyOptionFunc func() (x509.VerifyOptions, bool)

type Authenticator struct {
	verifyOptionsFn VerifyOptionFunc
	user            UserConversion
}

func New(opts x509.VerifyOptions, user UserConversion) *Authenticator {
	return NewDynamic(StaticVerifierFn(opts), user)
}

func NewDynamic(verifyOptionsFn VerifyOptionFunc, user UserConversion) *Authenticator {
	return &Authenticator{verifyOptionsFn, user}
}

func (a *Authenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	panic("not implemented")
}

type Verifier struct {
	verifyOptionsFn VerifyOptionFunc
	auth            authenticator.Request

	allowedCommonNames StringSliceProvider
}

func NewVerifier(opts x509.VerifyOptions, auth authenticator.Request, allowedCommonNames sets.String) authenticator.Request {
	return NewDynamicCAVerifier(StaticVerifierFn(opts), auth, StaticStringSlice(allowedCommonNames.List()))
}

func NewDynamicCAVerifier(verifyOptionsFn VerifyOptionFunc, auth authenticator.Request, allowedCommonNames StringSliceProvider) authenticator.Request {
	return &Verifier{verifyOptionsFn, auth, allowedCommonNames}
}

func (a *Verifier) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	panic("not implemented")
}

func (a *Verifier) verifySubject(subject pkix.Name) error {
	// No CN restrictions
	if len(a.allowedCommonNames.Value()) == 0 {
		return nil
	}
	// Enforce CN restrictions
	for _, allowedCommonName := range a.allowedCommonNames.Value() {
		if allowedCommonName == subject.CommonName {
			return nil
		}
	}
	return fmt.Errorf("x509: subject with cn=%s is not in the allowed list", subject.CommonName)
}

func DefaultVerifyOptions() x509.VerifyOptions {
	return x509.VerifyOptions{
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
}

var CommonNameUserConversion = UserConversionFunc(func(chain []*x509.Certificate) (*authenticator.Response, bool, error) {
	if len(chain[0].Subject.CommonName) == 0 {
		return nil, false, nil
	}
	return &authenticator.Response{
		User: &user.DefaultInfo{
			Name:   chain[0].Subject.CommonName,
			Groups: chain[0].Subject.Organization,
		},
	}, true, nil
})
