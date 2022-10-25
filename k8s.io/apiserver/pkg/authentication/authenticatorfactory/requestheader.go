package authenticatorfactory

import (
	"k8s.io/apiserver/pkg/authentication/request/headerrequest"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
)

type RequestHeaderConfig struct {
	// UsernameHeaders are the headers to check (in order, case-insensitively) for an identity. The first header with a value wins.
	UsernameHeaders headerrequest.StringSliceProvider
	// GroupHeaders are the headers to check (case-insensitively) for a group names.  All values will be used.
	GroupHeaders headerrequest.StringSliceProvider
	// ExtraHeaderPrefixes are the head prefixes to check (case-insentively) for filling in
	// the user.Info.Extra.  All values of all matching headers will be added.
	ExtraHeaderPrefixes headerrequest.StringSliceProvider
	// CAContentProvider the options for verifying incoming connections using mTLS.  Generally this points to CA bundle file which is used verify the identity of the front proxy.
	//	It may produce different options at will.
	CAContentProvider dynamiccertificates.CAContentProvider
	// AllowedClientNames is a list of common names that may be presented by the authenticating front proxy.  Empty means: accept any.
	AllowedClientNames headerrequest.StringSliceProvider
}
