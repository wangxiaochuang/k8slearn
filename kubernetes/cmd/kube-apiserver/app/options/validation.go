package options

import (
	"errors"
	"fmt"
	"net"
	"strings"

	apiextensionsapiserver "k8s.io/apiextensions-apiserver/pkg/apiserver"
	aggregatorscheme "k8s.io/kube-aggregator/pkg/apiserver/scheme"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	genericfeatures "k8s.io/apiserver/pkg/features"
	utilfeature "k8s.io/apiserver/pkg/util/feature"

	netutils "k8s.io/utils/net"
)

func validateClusterIPFlags(options *ServerRunOptions) []error {
	var errs []error
	const maxCIDRBits = 20

	if options.PrimaryServiceClusterIPRange.IP == nil {
		errs = append(errs, errors.New("--service-cluster-ip-range must contain at least one valid cidr"))
	}

	serviceClusterIPRangeList := strings.Split(options.ServiceClusterIPRanges, ",")
	if len(serviceClusterIPRangeList) > 2 {
		errs = append(errs, errors.New("--service-cluster-ip-range must not contain more than two entries"))
	}

	if err := validateMaxCIDRRange(options.PrimaryServiceClusterIPRange, maxCIDRBits, "--service-cluster-ip-range"); err != nil {
		errs = append(errs, err)
	}

	secondaryServiceClusterIPRangeUsed := (options.SecondaryServiceClusterIPRange.IP != nil)
	if secondaryServiceClusterIPRangeUsed {
		dualstack, err := netutils.IsDualStackCIDRs([]*net.IPNet{&options.PrimaryServiceClusterIPRange, &options.SecondaryServiceClusterIPRange})
		if err != nil {
			errs = append(errs, fmt.Errorf("error attempting to validate dualstack for --service-cluster-ip-range value error:%v", err))
		}

		if !dualstack {
			errs = append(errs, errors.New("--service-cluster-ip-range[0] and --service-cluster-ip-range[1] must be of different IP family"))
		}

		if err := validateMaxCIDRRange(options.SecondaryServiceClusterIPRange, maxCIDRBits, "--service-cluster-ip-range[1]"); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func validateMaxCIDRRange(cidr net.IPNet, maxCIDRBits int, cidrFlag string) error {
	var ones, bits = cidr.Mask.Size()
	if bits-ones > maxCIDRBits {
		return fmt.Errorf("specified %s is too large; for %d-bit addresses, the mask must be >= %d", cidrFlag, bits, bits-maxCIDRBits)
	}

	return nil
}

func validateServiceNodePort(options *ServerRunOptions) []error {
	var errs []error

	if options.KubernetesServiceNodePort < 0 || options.KubernetesServiceNodePort > 65535 {
		errs = append(errs, fmt.Errorf("--kubernetes-service-node-port %v must be between 0 and 65535, inclusive. If 0, the Kubernetes master service will be of type ClusterIP", options.KubernetesServiceNodePort))
	}

	if options.KubernetesServiceNodePort > 0 && !options.ServiceNodePortRange.Contains(options.KubernetesServiceNodePort) {
		errs = append(errs, fmt.Errorf("kubernetes service node port range %v doesn't contain %v", options.ServiceNodePortRange, options.KubernetesServiceNodePort))
	}
	return errs
}

func validateAPIPriorityAndFairness(options *ServerRunOptions) []error {
	if utilfeature.DefaultFeatureGate.Enabled(genericfeatures.APIPriorityAndFairness) && options.GenericServerRunOptions.EnablePriorityAndFairness {
		// If none of the following runtime config options are specified, APF is
		// assumed to be turned on.
		enabledAPIString := options.APIEnablement.RuntimeConfig.String()
		testConfigs := []string{"flowcontrol.apiserver.k8s.io/v1beta2", "flowcontrol.apiserver.k8s.io/v1beta1", "api/beta", "api/all"} // in the order of precedence
		for _, testConfig := range testConfigs {
			if strings.Contains(enabledAPIString, fmt.Sprintf("%s=false", testConfig)) {
				return []error{fmt.Errorf("--runtime-config=%s=false conflicts with --enable-priority-and-fairness=true and --feature-gates=APIPriorityAndFairness=true", testConfig)}
			}
			if strings.Contains(enabledAPIString, fmt.Sprintf("%s=true", testConfig)) {
				return nil
			}
		}
	}

	return nil
}

func (s *ServerRunOptions) Validate() []error {
	var errs []error
	if s.MasterCount <= 0 {
		errs = append(errs, fmt.Errorf("--apiserver-count should be a positive number, but value '%d' provided", s.MasterCount))
	}
	errs = append(errs, s.Etcd.Validate()...)
	errs = append(errs, validateClusterIPFlags(s)...)
	errs = append(errs, validateServiceNodePort(s)...)
	errs = append(errs, validateAPIPriorityAndFairness(s)...)
	errs = append(errs, s.SecureServing.Validate()...)
	errs = append(errs, s.Authentication.Validate()...)
	errs = append(errs, s.Authorization.Validate()...)
	errs = append(errs, s.Audit.Validate()...)
	errs = append(errs, s.APIEnablement.Validate(legacyscheme.Scheme, apiextensionsapiserver.Scheme, aggregatorscheme.Scheme)...)

	return errs
}
