package audit

import (
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

type RequestAuditConfig struct {
	OmitStages []audit.Stage

	OmitManagedFields bool
}

type RequestAuditConfigWithLevel struct {
	RequestAuditConfig

	Level audit.Level
}

type PolicyRuleEvaluator interface {
	EvaluatePolicyRule(authorizer.Attributes) RequestAuditConfigWithLevel
}
