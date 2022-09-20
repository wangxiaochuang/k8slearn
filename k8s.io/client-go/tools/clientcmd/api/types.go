package api

import "k8s.io/apimachinery/pkg/runtime"

// p171
type AuthProviderConfig struct {
	Name   string            `json:"name"`
	Config map[string]string `json:"config,omitempty"`
}

// p201
type ExecConfig struct {
	Command                 string       `json:"command"`
	Args                    []string     `json:"args"`
	Env                     []ExecEnvVar `json:"env"`
	APIVersion              string       `json:"apiVersion,omitempty"`
	InstallHint             string       `json:"installHint,omitempty"`
	ProvideClusterInfo      bool         `json:"provideClusterInfo"`
	Config                  runtime.Object
	InteractiveMode         ExecInteractiveMode
	StdinUnavailable        bool
	StdinUnavailableMessage string
}

// p309
type ExecEnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// p315
type ExecInteractiveMode string
