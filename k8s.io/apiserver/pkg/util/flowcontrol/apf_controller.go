package flowcontrol

import (
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
)

// p91
type RequestDigest struct {
	RequestInfo *request.RequestInfo
	User        user.Info
}
