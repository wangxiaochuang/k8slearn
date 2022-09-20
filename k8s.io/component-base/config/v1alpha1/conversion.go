package v1alpha1

import (
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/component-base/config"
)

func Convert_v1alpha1_ClientConnectionConfiguration_To_config_ClientConnectionConfiguration(in *ClientConnectionConfiguration, out *config.ClientConnectionConfiguration, s conversion.Scope) error {
	return autoConvert_v1alpha1_ClientConnectionConfiguration_To_config_ClientConnectionConfiguration(in, out, s)
}

func Convert_config_ClientConnectionConfiguration_To_v1alpha1_ClientConnectionConfiguration(in *config.ClientConnectionConfiguration, out *ClientConnectionConfiguration, s conversion.Scope) error {
	return autoConvert_config_ClientConnectionConfiguration_To_v1alpha1_ClientConnectionConfiguration(in, out, s)
}

func Convert_v1alpha1_DebuggingConfiguration_To_config_DebuggingConfiguration(in *DebuggingConfiguration, out *config.DebuggingConfiguration, s conversion.Scope) error {
	return autoConvert_v1alpha1_DebuggingConfiguration_To_config_DebuggingConfiguration(in, out, s)
}

func Convert_config_DebuggingConfiguration_To_v1alpha1_DebuggingConfiguration(in *config.DebuggingConfiguration, out *DebuggingConfiguration, s conversion.Scope) error {
	return autoConvert_config_DebuggingConfiguration_To_v1alpha1_DebuggingConfiguration(in, out, s)
}

func Convert_v1alpha1_LeaderElectionConfiguration_To_config_LeaderElectionConfiguration(in *LeaderElectionConfiguration, out *config.LeaderElectionConfiguration, s conversion.Scope) error {
	return autoConvert_v1alpha1_LeaderElectionConfiguration_To_config_LeaderElectionConfiguration(in, out, s)
}

func Convert_config_LeaderElectionConfiguration_To_v1alpha1_LeaderElectionConfiguration(in *config.LeaderElectionConfiguration, out *LeaderElectionConfiguration, s conversion.Scope) error {
	return autoConvert_config_LeaderElectionConfiguration_To_v1alpha1_LeaderElectionConfiguration(in, out, s)
}

func Convert_v1alpha1_LoggingConfiguration_To_config_LoggingConfiguration(in *LoggingConfiguration, out *config.LoggingConfiguration, s conversion.Scope) error {
	return autoConvert_v1alpha1_LoggingConfiguration_To_config_LoggingConfiguration(in, out, s)
}

func Convert_config_LoggingConfiguration_To_v1alpha1_LoggingConfiguration(in *config.LoggingConfiguration, out *LoggingConfiguration, s conversion.Scope) error {
	return autoConvert_config_LoggingConfiguration_To_v1alpha1_LoggingConfiguration(in, out, s)
}
