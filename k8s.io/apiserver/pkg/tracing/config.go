package tracing

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/apis/apiserver"
	"k8s.io/apiserver/pkg/apis/apiserver/install"
)

const (
	maxSamplingRatePerMillion = 1000000
)

var (
	cfgScheme = runtime.NewScheme()
	codecs    = serializer.NewCodecFactory(cfgScheme)
)

func init() {
	install.Install(cfgScheme)
}

func ReadTracingConfiguration(configFilePath string) (*apiserver.TracingConfiguration, error) {
	if configFilePath == "" {
		return nil, fmt.Errorf("tracing config file was empty")
	}
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read tracing configuration from %q: %v", configFilePath, err)
	}
	internalConfig := &apiserver.TracingConfiguration{}
	// 调用codecs的Decode方法，将byte数组data转换为内部版本结构体
	if err := runtime.DecodeInto(codecs.UniversalDecoder(), data, internalConfig); err != nil {
		return nil, fmt.Errorf("unable to decode tracing configuration data: %v", err)
	}
	return internalConfig, nil
}

func ValidateTracingConfiguration(config *apiserver.TracingConfiguration) field.ErrorList {
	allErrs := field.ErrorList{}
	if config == nil {
		// Tracing is disabled
		return allErrs
	}
	if config.SamplingRatePerMillion != nil {
		allErrs = append(allErrs, validateSamplingRate(*config.SamplingRatePerMillion, field.NewPath("samplingRatePerMillion"))...)
	}
	if config.Endpoint != nil {
		allErrs = append(allErrs, validateEndpoint(*config.Endpoint, field.NewPath("endpoint"))...)
	}
	return allErrs
}

func validateSamplingRate(rate int32, fldPath *field.Path) field.ErrorList {
	errs := field.ErrorList{}
	if rate < 0 {
		errs = append(errs, field.Invalid(
			fldPath, rate,
			"sampling rate must be positive",
		))
	}
	if rate > maxSamplingRatePerMillion {
		errs = append(errs, field.Invalid(
			fldPath, rate,
			"sampling rate per million must be less than or equal to one million",
		))
	}
	return errs
}

func validateEndpoint(endpoint string, fldPath *field.Path) field.ErrorList {
	errs := field.ErrorList{}
	if !strings.Contains(endpoint, "//") {
		endpoint = "dns://" + endpoint
	}
	url, err := url.Parse(endpoint)
	if err != nil {
		errs = append(errs, field.Invalid(
			fldPath, endpoint,
			err.Error(),
		))
		return errs
	}
	switch url.Scheme {
	case "dns":
	case "unix":
	case "unix-abstract":
	default:
		errs = append(errs, field.Invalid(
			fldPath, endpoint,
			fmt.Sprintf("unsupported scheme: %v.  Options are none, dns, unix, or unix-abstract.  See https://github.com/grpc/grpc/blob/master/doc/naming.md", url.Scheme),
		))
	}
	return errs
}
