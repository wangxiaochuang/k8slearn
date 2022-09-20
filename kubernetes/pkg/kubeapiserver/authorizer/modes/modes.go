package models

import "k8s.io/apimachinery/pkg/util/sets"

const (
	// ModeAlwaysAllow is the mode to set all requests as authorized
	ModeAlwaysAllow string = "AlwaysAllow"
	// ModeAlwaysDeny is the mode to set no requests as authorized
	ModeAlwaysDeny string = "AlwaysDeny"
	// ModeABAC is the mode to use Attribute Based Access Control to authorize
	ModeABAC string = "ABAC"
	// ModeWebhook is the mode to make an external webhook call to authorize
	ModeWebhook string = "Webhook"
	// ModeRBAC is the mode to use Role Based Access Control to authorize
	ModeRBAC string = "RBAC"
	// ModeNode is an authorization mode that authorizes API requests made by kubelets.
	ModeNode string = "Node"
)

var AuthorizationModeChoices = []string{ModeAlwaysAllow, ModeAlwaysDeny, ModeABAC, ModeWebhook, ModeRBAC, ModeNode}

// IsValidAuthorizationMode returns true if the given authorization mode is a valid one for the apiserver
func IsValidAuthorizationMode(authzMode string) bool {
	return sets.NewString(AuthorizationModeChoices...).Has(authzMode)
}
