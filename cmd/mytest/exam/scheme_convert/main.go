package scheme_convert

import (
	"fmt"
	"io/ioutil"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/apis/apiserver"
	"k8s.io/apiserver/pkg/apis/apiserver/v1beta1"
	"k8s.io/utils/wxc"
	"sigs.k8s.io/yaml"
)

var cfgScheme = runtime.NewScheme()

func init() {
	apiserver.AddToScheme(cfgScheme)
	v1beta1.AddToScheme(cfgScheme)
}

func main() {
	data, _ := ioutil.ReadFile("testdata/kube_egress_selector_configuration.yaml")

	var decodedConfig v1beta1.EgressSelectorConfiguration
	err := yaml.Unmarshal(data, &decodedConfig)
	if err != nil {
		// we got an error where the decode wasn't related to a missing type
		return
	}
	internalConfig := &apiserver.EgressSelectorConfiguration{}

	if err := cfgScheme.Convert(&decodedConfig, internalConfig, nil); err != nil {
		fmt.Println(err)
		return
	}
	wxc.P(internalConfig)
}
