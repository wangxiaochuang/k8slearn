package register

import (
	"k8s.io/component-base/logs"
	json "k8s.io/component-base/logs/json"
	"k8s.io/component-base/logs/registry"
)

func init() {
	registry.LogRegistry.Register(logs.JSONLogFormat, json.Factory{})
}
