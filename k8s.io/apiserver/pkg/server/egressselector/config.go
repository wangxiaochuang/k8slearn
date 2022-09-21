package egressselector

import (
	"fmt"
	"io/ioutil"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/apiserver/pkg/apis/apiserver"
	"k8s.io/apiserver/pkg/apis/apiserver/install"
	"k8s.io/apiserver/pkg/apis/apiserver/v1beta1"
)

var cfgScheme = runtime.NewScheme()

var validEgressSelectorNames = sets.NewString("controlplane", "cluster", "etcd")

func init() {
	install.Install(cfgScheme)
}

func ReadEgressSelectorConfiguration(configFilePath string) (*apiserver.EgressSelectorConfiguration, error) {
	if configFilePath == "" {
		return nil, nil
	}

	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read egress selector configuration from %q [%v]", configFilePath, err)
	}
	var decodedConfig v1beta1.EgressSelectorConfiguration
	err = yaml.Unmarshal(data, &decodedConfig)
	if err != nil {
		// we got an error where the decode wasn't related to a missing type
		return nil, err
	}
	if decodedConfig.Kind != "EgressSelectorConfiguration" {
		return nil, fmt.Errorf("invalid service configuration object %q", decodedConfig.Kind)
	}
	internalConfig := &apiserver.EgressSelectorConfiguration{}
	if err := cfgScheme.Convert(&decodedConfig, internalConfig, nil); err != nil {
		// we got an error where the decode wasn't related to a missing type
		return nil, err
	}
	return internalConfig, nil
}
