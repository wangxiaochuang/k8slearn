package request

import "net/http"

type LongRunningRequestCheck func(r *http.Request, requestInfo *RequestInfo) bool

type RequestInfoResolver interface {
	NewRequestInfo(req *http.Request) (*RequestInfo, error)
}

type RequestInfo struct {
	IsResourceRequest bool
	Path              string
	Verb              string

	APIPrefix   string
	APIGroup    string
	APIVersion  string
	Namespace   string
	Resource    string
	Subresource string
	Name        string
	Parts       []string
}
