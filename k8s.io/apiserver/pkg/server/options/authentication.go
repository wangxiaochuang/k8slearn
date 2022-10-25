package options

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/authentication/authenticatorfactory"
	"k8s.io/apiserver/pkg/authentication/request/headerrequest"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
)

func DefaultAuthWebhookRetryBackoff() *wait.Backoff {
	return &wait.Backoff{
		Duration: 500 * time.Millisecond,
		Factor:   1.5,
		Jitter:   0.2,
		Steps:    5,
	}
}

// p52
type RequestHeaderAuthenticationOptions struct {
	ClientCAFile string

	UsernameHeaders     []string
	GroupHeaders        []string
	ExtraHeaderPrefixes []string
	AllowedNames        []string
}

func (s *RequestHeaderAuthenticationOptions) Validate() []error {
	allErrors := []error{}

	if err := checkForWhiteSpaceOnly("requestheader-username-headers", s.UsernameHeaders...); err != nil {
		allErrors = append(allErrors, err)
	}
	if err := checkForWhiteSpaceOnly("requestheader-group-headers", s.GroupHeaders...); err != nil {
		allErrors = append(allErrors, err)
	}
	if err := checkForWhiteSpaceOnly("requestheader-extra-headers-prefix", s.ExtraHeaderPrefixes...); err != nil {
		allErrors = append(allErrors, err)
	}
	if err := checkForWhiteSpaceOnly("requestheader-allowed-names", s.AllowedNames...); err != nil {
		allErrors = append(allErrors, err)
	}

	return allErrors
}

func checkForWhiteSpaceOnly(flag string, headerNames ...string) error {
	for _, headerName := range headerNames {
		if len(strings.TrimSpace(headerName)) == 0 {
			return fmt.Errorf("empty value in %q", flag)
		}
	}

	return nil
}

func (s *RequestHeaderAuthenticationOptions) AddFlags(fs *pflag.FlagSet) {
	if s == nil {
		return
	}

	fs.StringSliceVar(&s.UsernameHeaders, "requestheader-username-headers", s.UsernameHeaders, ""+
		"List of request headers to inspect for usernames. X-Remote-User is common.")

	fs.StringSliceVar(&s.GroupHeaders, "requestheader-group-headers", s.GroupHeaders, ""+
		"List of request headers to inspect for groups. X-Remote-Group is suggested.")

	fs.StringSliceVar(&s.ExtraHeaderPrefixes, "requestheader-extra-headers-prefix", s.ExtraHeaderPrefixes, ""+
		"List of request header prefixes to inspect. X-Remote-Extra- is suggested.")

	fs.StringVar(&s.ClientCAFile, "requestheader-client-ca-file", s.ClientCAFile, ""+
		"Root certificate bundle to use to verify client certificates on incoming requests "+
		"before trusting usernames in headers specified by --requestheader-username-headers. "+
		"WARNING: generally do not depend on authorization being already done for incoming requests.")

	fs.StringSliceVar(&s.AllowedNames, "requestheader-allowed-names", s.AllowedNames, ""+
		"List of client certificate common names to allow to provide usernames in headers "+
		"specified by --requestheader-username-headers. If empty, any client certificate validated "+
		"by the authorities in --requestheader-client-ca-file is allowed.")
}

func (s *RequestHeaderAuthenticationOptions) ToAuthenticationRequestHeaderConfig() (*authenticatorfactory.RequestHeaderConfig, error) {
	if len(s.ClientCAFile) == 0 {
		return nil, nil
	}

	caBundleProvider, err := dynamiccertificates.NewDynamicCAContentFromFile("request-header", s.ClientCAFile)
	if err != nil {
		return nil, err
	}

	return &authenticatorfactory.RequestHeaderConfig{
		UsernameHeaders:     headerrequest.StaticStringSlice(s.UsernameHeaders),
		GroupHeaders:        headerrequest.StaticStringSlice(s.GroupHeaders),
		ExtraHeaderPrefixes: headerrequest.StaticStringSlice(s.ExtraHeaderPrefixes),
		CAContentProvider:   caBundleProvider,
		AllowedClientNames:  headerrequest.StaticStringSlice(s.AllowedNames),
	}, nil
}

// p140
type ClientCertAuthenticationOptions struct {
	ClientCA          string
	CAContentProvider dynamiccertificates.CAContentProvider
}

// p151
func (s *ClientCertAuthenticationOptions) GetClientCAContentProvider() (dynamiccertificates.CAContentProvider, error) {
	if s.CAContentProvider != nil {
		return s.CAContentProvider, nil
	}

	if len(s.ClientCA) == 0 {
		return nil, nil
	}

	return dynamiccertificates.NewDynamicCAContentFromFile("client-ca-bundle", s.ClientCA)
}

// 163
func (s *ClientCertAuthenticationOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.ClientCA, "client-ca-file", s.ClientCA, ""+
		"If set, any request presenting a client certificate signed by one of "+
		"the authorities in the client-ca-file is authenticated with an identity "+
		"corresponding to the CommonName of the client certificate.")
}
