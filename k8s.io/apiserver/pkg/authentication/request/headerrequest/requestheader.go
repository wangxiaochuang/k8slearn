package headerrequest

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	x509request "k8s.io/apiserver/pkg/authentication/request/x509"
	utilcert "k8s.io/client-go/util/cert"
)

type StringSliceProvider interface {
	Value() []string
}

type StringSliceProviderFunc func() []string

func (d StringSliceProviderFunc) Value() []string {
	return d()
}

type StaticStringSlice []string

func (s StaticStringSlice) Value() []string {
	return s
}

type requestHeaderAuthRequestHandler struct {
	nameHeaders         StringSliceProvider
	groupHeaders        StringSliceProvider
	extraHeaderPrefixes StringSliceProvider
}

func New(nameHeaders, groupHeaders, extraHeaderPrefixes []string) (authenticator.Request, error) {
	trimmedNameHeaders, err := trimHeaders(nameHeaders...)
	if err != nil {
		return nil, err
	}
	trimmedGroupHeaders, err := trimHeaders(groupHeaders...)
	if err != nil {
		return nil, err
	}
	trimmedExtraHeaderPrefixes, err := trimHeaders(extraHeaderPrefixes...)
	if err != nil {
		return nil, err
	}

	return NewDynamic(
		StaticStringSlice(trimmedNameHeaders),
		StaticStringSlice(trimmedGroupHeaders),
		StaticStringSlice(trimmedExtraHeaderPrefixes),
	), nil
}

func NewDynamic(nameHeaders, groupHeaders, extraHeaderPrefixes StringSliceProvider) authenticator.Request {
	return &requestHeaderAuthRequestHandler{
		nameHeaders:         nameHeaders,
		groupHeaders:        groupHeaders,
		extraHeaderPrefixes: extraHeaderPrefixes,
	}
}

func trimHeaders(headerNames ...string) ([]string, error) {
	ret := []string{}
	for _, headerName := range headerNames {
		trimmedHeader := strings.TrimSpace(headerName)
		if len(trimmedHeader) == 0 {
			return nil, fmt.Errorf("empty header %q", headerName)
		}
		ret = append(ret, trimmedHeader)
	}

	return ret, nil
}

func NewSecure(clientCA string, proxyClientNames []string, nameHeaders []string, groupHeaders []string, extraHeaderPrefixes []string) (authenticator.Request, error) {
	if len(clientCA) == 0 {
		return nil, fmt.Errorf("missing clientCA file")
	}

	// Wrap with an x509 verifier
	caData, err := ioutil.ReadFile(clientCA)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %v", clientCA, err)
	}
	opts := x509request.DefaultVerifyOptions()
	opts.Roots = x509.NewCertPool()
	certs, err := utilcert.ParseCertsPEM(caData)
	if err != nil {
		return nil, fmt.Errorf("error loading certs from  %s: %v", clientCA, err)
	}
	for _, cert := range certs {
		opts.Roots.AddCert(cert)
	}

	trimmedNameHeaders, err := trimHeaders(nameHeaders...)
	if err != nil {
		return nil, err
	}
	trimmedGroupHeaders, err := trimHeaders(groupHeaders...)
	if err != nil {
		return nil, err
	}
	trimmedExtraHeaderPrefixes, err := trimHeaders(extraHeaderPrefixes...)
	if err != nil {
		return nil, err
	}

	return NewDynamicVerifyOptionsSecure(
		x509request.StaticVerifierFn(opts),
		StaticStringSlice(proxyClientNames),
		StaticStringSlice(trimmedNameHeaders),
		StaticStringSlice(trimmedGroupHeaders),
		StaticStringSlice(trimmedExtraHeaderPrefixes),
	), nil
}

func NewDynamicVerifyOptionsSecure(verifyOptionFn x509request.VerifyOptionFunc, proxyClientNames, nameHeaders, groupHeaders, extraHeaderPrefixes StringSliceProvider) authenticator.Request {
	headerAuthenticator := NewDynamic(nameHeaders, groupHeaders, extraHeaderPrefixes)

	return x509request.NewDynamicCAVerifier(verifyOptionFn, headerAuthenticator, proxyClientNames)
}

func (a *requestHeaderAuthRequestHandler) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	panic("not implemented")
}

func headerValue(h http.Header, headerNames []string) string {
	for _, headerName := range headerNames {
		headerValue := h.Get(headerName)
		if len(headerValue) > 0 {
			return headerValue
		}
	}
	return ""
}

func allHeaderValues(h http.Header, headerNames []string) []string {
	ret := []string{}
	for _, headerName := range headerNames {
		headerKey := http.CanonicalHeaderKey(headerName)
		values, ok := h[headerKey]
		if !ok {
			continue
		}

		for _, headerValue := range values {
			if len(headerValue) > 0 {
				ret = append(ret, headerValue)
			}
		}
	}
	return ret
}

func unescapeExtraKey(encodedKey string) string {
	key, err := url.PathUnescape(encodedKey) // Decode %-encoded bytes.
	if err != nil {
		return encodedKey // Always record extra strings, even if malformed/unencoded.
	}
	return key
}

func newExtra(h http.Header, headerPrefixes []string) map[string][]string {
	ret := map[string][]string{}

	// we have to iterate over prefixes first in order to have proper ordering inside the value slices
	for _, prefix := range headerPrefixes {
		for headerName, vv := range h {
			if !strings.HasPrefix(strings.ToLower(headerName), strings.ToLower(prefix)) {
				continue
			}

			extraKey := unescapeExtraKey(strings.ToLower(headerName[len(prefix):]))
			ret[extraKey] = append(ret[extraKey], vv...)
		}
	}

	return ret
}
