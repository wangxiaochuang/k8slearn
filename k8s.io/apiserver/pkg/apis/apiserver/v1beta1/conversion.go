package v1beta1

import (
	conversion "k8s.io/apimachinery/pkg/conversion"
	apiserver "k8s.io/apiserver/pkg/apis/apiserver"
)

func Convert_v1beta1_EgressSelection_To_apiserver_EgressSelection(in *EgressSelection, out *apiserver.EgressSelection, s conversion.Scope) error {
	if err := autoConvert_v1beta1_EgressSelection_To_apiserver_EgressSelection(in, out, s); err != nil {
		return err
	}
	if out.Name == "master" {
		out.Name = "controlplane"
	}
	return nil
}
