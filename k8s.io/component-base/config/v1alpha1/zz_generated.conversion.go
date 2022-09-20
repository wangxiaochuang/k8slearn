package v1alpha1

import (
	"time"
	"unsafe"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/component-base/config"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// 将各种类型间的转换函数放到scheme里
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*FormatOptions)(nil), (*config.FormatOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_FormatOptions_To_config_FormatOptions(a.(*FormatOptions), b.(*config.FormatOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*config.FormatOptions)(nil), (*FormatOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_config_FormatOptions_To_v1alpha1_FormatOptions(a.(*config.FormatOptions), b.(*FormatOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*JSONOptions)(nil), (*config.JSONOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_JSONOptions_To_config_JSONOptions(a.(*JSONOptions), b.(*config.JSONOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*config.JSONOptions)(nil), (*JSONOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_config_JSONOptions_To_v1alpha1_JSONOptions(a.(*config.JSONOptions), b.(*JSONOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VModuleItem)(nil), (*config.VModuleItem)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VModuleItem_To_config_VModuleItem(a.(*VModuleItem), b.(*config.VModuleItem), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*config.VModuleItem)(nil), (*VModuleItem)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_config_VModuleItem_To_v1alpha1_VModuleItem(a.(*config.VModuleItem), b.(*VModuleItem), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*config.ClientConnectionConfiguration)(nil), (*ClientConnectionConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_config_ClientConnectionConfiguration_To_v1alpha1_ClientConnectionConfiguration(a.(*config.ClientConnectionConfiguration), b.(*ClientConnectionConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*config.DebuggingConfiguration)(nil), (*DebuggingConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_config_DebuggingConfiguration_To_v1alpha1_DebuggingConfiguration(a.(*config.DebuggingConfiguration), b.(*DebuggingConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*config.LeaderElectionConfiguration)(nil), (*LeaderElectionConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_config_LeaderElectionConfiguration_To_v1alpha1_LeaderElectionConfiguration(a.(*config.LeaderElectionConfiguration), b.(*LeaderElectionConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*config.LoggingConfiguration)(nil), (*LoggingConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_config_LoggingConfiguration_To_v1alpha1_LoggingConfiguration(a.(*config.LoggingConfiguration), b.(*LoggingConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*ClientConnectionConfiguration)(nil), (*config.ClientConnectionConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_ClientConnectionConfiguration_To_config_ClientConnectionConfiguration(a.(*ClientConnectionConfiguration), b.(*config.ClientConnectionConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*DebuggingConfiguration)(nil), (*config.DebuggingConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_DebuggingConfiguration_To_config_DebuggingConfiguration(a.(*DebuggingConfiguration), b.(*config.DebuggingConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*LeaderElectionConfiguration)(nil), (*config.LeaderElectionConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_LeaderElectionConfiguration_To_config_LeaderElectionConfiguration(a.(*LeaderElectionConfiguration), b.(*config.LeaderElectionConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*LoggingConfiguration)(nil), (*config.LoggingConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_LoggingConfiguration_To_config_LoggingConfiguration(a.(*LoggingConfiguration), b.(*config.LoggingConfiguration), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_ClientConnectionConfiguration_To_config_ClientConnectionConfiguration(in *ClientConnectionConfiguration, out *config.ClientConnectionConfiguration, s conversion.Scope) error {
	out.Kubeconfig = in.Kubeconfig
	out.AcceptContentTypes = in.AcceptContentTypes
	out.ContentType = in.ContentType
	out.QPS = in.QPS
	out.Burst = in.Burst
	return nil
}

func autoConvert_config_ClientConnectionConfiguration_To_v1alpha1_ClientConnectionConfiguration(in *config.ClientConnectionConfiguration, out *ClientConnectionConfiguration, s conversion.Scope) error {
	out.Kubeconfig = in.Kubeconfig
	out.AcceptContentTypes = in.AcceptContentTypes
	out.ContentType = in.ContentType
	out.QPS = in.QPS
	out.Burst = in.Burst
	return nil
}

func autoConvert_v1alpha1_DebuggingConfiguration_To_config_DebuggingConfiguration(in *DebuggingConfiguration, out *config.DebuggingConfiguration, s conversion.Scope) error {
	if err := v1.Convert_Pointer_bool_To_bool(&in.EnableProfiling, &out.EnableProfiling, s); err != nil {
		return err
	}
	if err := v1.Convert_Pointer_bool_To_bool(&in.EnableContentionProfiling, &out.EnableContentionProfiling, s); err != nil {
		return err
	}
	return nil
}

func autoConvert_config_DebuggingConfiguration_To_v1alpha1_DebuggingConfiguration(in *config.DebuggingConfiguration, out *DebuggingConfiguration, s conversion.Scope) error {
	if err := v1.Convert_bool_To_Pointer_bool(&in.EnableProfiling, &out.EnableProfiling, s); err != nil {
		return err
	}
	if err := v1.Convert_bool_To_Pointer_bool(&in.EnableContentionProfiling, &out.EnableContentionProfiling, s); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_FormatOptions_To_config_FormatOptions(in *FormatOptions, out *config.FormatOptions, s conversion.Scope) error {
	if err := Convert_v1alpha1_JSONOptions_To_config_JSONOptions(&in.JSON, &out.JSON, s); err != nil {
		return err
	}
	return nil
}

func Convert_v1alpha1_FormatOptions_To_config_FormatOptions(in *FormatOptions, out *config.FormatOptions, s conversion.Scope) error {
	return autoConvert_v1alpha1_FormatOptions_To_config_FormatOptions(in, out, s)
}

func autoConvert_config_FormatOptions_To_v1alpha1_FormatOptions(in *config.FormatOptions, out *FormatOptions, s conversion.Scope) error {
	if err := Convert_config_JSONOptions_To_v1alpha1_JSONOptions(&in.JSON, &out.JSON, s); err != nil {
		return err
	}
	return nil
}

func Convert_config_FormatOptions_To_v1alpha1_FormatOptions(in *config.FormatOptions, out *FormatOptions, s conversion.Scope) error {
	return autoConvert_config_FormatOptions_To_v1alpha1_FormatOptions(in, out, s)
}

func autoConvert_v1alpha1_JSONOptions_To_config_JSONOptions(in *JSONOptions, out *config.JSONOptions, s conversion.Scope) error {
	out.SplitStream = in.SplitStream
	out.InfoBufferSize = in.InfoBufferSize
	return nil
}

func Convert_v1alpha1_JSONOptions_To_config_JSONOptions(in *JSONOptions, out *config.JSONOptions, s conversion.Scope) error {
	return autoConvert_v1alpha1_JSONOptions_To_config_JSONOptions(in, out, s)
}

func autoConvert_config_JSONOptions_To_v1alpha1_JSONOptions(in *config.JSONOptions, out *JSONOptions, s conversion.Scope) error {
	out.SplitStream = in.SplitStream
	out.InfoBufferSize = in.InfoBufferSize
	return nil
}

func Convert_config_JSONOptions_To_v1alpha1_JSONOptions(in *config.JSONOptions, out *JSONOptions, s conversion.Scope) error {
	return autoConvert_config_JSONOptions_To_v1alpha1_JSONOptions(in, out, s)
}

func autoConvert_v1alpha1_LeaderElectionConfiguration_To_config_LeaderElectionConfiguration(in *LeaderElectionConfiguration, out *config.LeaderElectionConfiguration, s conversion.Scope) error {
	if err := v1.Convert_Pointer_bool_To_bool(&in.LeaderElect, &out.LeaderElect, s); err != nil {
		return err
	}
	out.LeaseDuration = in.LeaseDuration
	out.RenewDeadline = in.RenewDeadline
	out.RetryPeriod = in.RetryPeriod
	out.ResourceLock = in.ResourceLock
	out.ResourceName = in.ResourceName
	out.ResourceNamespace = in.ResourceNamespace
	return nil
}

func autoConvert_config_LeaderElectionConfiguration_To_v1alpha1_LeaderElectionConfiguration(in *config.LeaderElectionConfiguration, out *LeaderElectionConfiguration, s conversion.Scope) error {
	if err := v1.Convert_bool_To_Pointer_bool(&in.LeaderElect, &out.LeaderElect, s); err != nil {
		return err
	}
	out.LeaseDuration = in.LeaseDuration
	out.RenewDeadline = in.RenewDeadline
	out.RetryPeriod = in.RetryPeriod
	out.ResourceLock = in.ResourceLock
	out.ResourceName = in.ResourceName
	out.ResourceNamespace = in.ResourceNamespace
	return nil
}

// p224
func autoConvert_v1alpha1_LoggingConfiguration_To_config_LoggingConfiguration(in *LoggingConfiguration, out *config.LoggingConfiguration, s conversion.Scope) error {
	out.Format = in.Format
	out.FlushFrequency = time.Duration(in.FlushFrequency)
	out.Verbosity = config.VerbosityLevel(in.Verbosity)
	out.VModule = *(*config.VModuleConfiguration)(unsafe.Pointer(&in.VModule))
	if err := Convert_v1alpha1_FormatOptions_To_config_FormatOptions(&in.Options, &out.Options, s); err != nil {
		return err
	}
	return nil
}

func autoConvert_config_LoggingConfiguration_To_v1alpha1_LoggingConfiguration(in *config.LoggingConfiguration, out *LoggingConfiguration, s conversion.Scope) error {
	out.Format = in.Format
	out.FlushFrequency = time.Duration(in.FlushFrequency)
	out.Verbosity = uint32(in.Verbosity)
	out.VModule = *(*VModuleConfiguration)(unsafe.Pointer(&in.VModule))
	if err := Convert_config_FormatOptions_To_v1alpha1_FormatOptions(&in.Options, &out.Options, s); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_VModuleItem_To_config_VModuleItem(in *VModuleItem, out *config.VModuleItem, s conversion.Scope) error {
	out.FilePattern = in.FilePattern
	out.Verbosity = config.VerbosityLevel(in.Verbosity)
	return nil
}

func Convert_v1alpha1_VModuleItem_To_config_VModuleItem(in *VModuleItem, out *config.VModuleItem, s conversion.Scope) error {
	return autoConvert_v1alpha1_VModuleItem_To_config_VModuleItem(in, out, s)
}

func autoConvert_config_VModuleItem_To_v1alpha1_VModuleItem(in *config.VModuleItem, out *VModuleItem, s conversion.Scope) error {
	out.FilePattern = in.FilePattern
	out.Verbosity = uint32(in.Verbosity)
	return nil
}

func Convert_config_VModuleItem_To_v1alpha1_VModuleItem(in *config.VModuleItem, out *VModuleItem, s conversion.Scope) error {
	return autoConvert_config_VModuleItem_To_v1alpha1_VModuleItem(in, out, s)
}
