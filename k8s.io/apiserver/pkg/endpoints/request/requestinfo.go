package request

import (
	"context"
	"net/http"
	"strings"
)

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

// p249
type requestInfoKeyType int

// p254
const requestInfoKey requestInfoKeyType = iota

// WithRequestInfo returns a copy of parent in which the request info value is set
func WithRequestInfo(parent context.Context, info *RequestInfo) context.Context {
	return WithValue(parent, requestInfoKey, info)
}

// p262
func RequestInfoFrom(ctx context.Context) (*RequestInfo, bool) {
	info, ok := ctx.Value(requestInfoKey).(*RequestInfo)
	return info, ok
}

// splitPath returns the segments for a URL path.
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}
