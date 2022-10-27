package options

import (
	"github.com/spf13/pflag"
)

// CloudProviderOptions contains cloud provider config
type CloudProviderOptions struct {
	CloudConfigFile string
	CloudProvider   string
}

// NewCloudProviderOptions creates a default CloudProviderOptions
func NewCloudProviderOptions() *CloudProviderOptions {
	return &CloudProviderOptions{}
}

// Validate checks invalid config
func (s *CloudProviderOptions) Validate() []error {
	allErrors := []error{}
	return allErrors
}

// AddFlags returns flags of cloud provider for a API Server
func (s *CloudProviderOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.CloudProvider, "cloud-provider", s.CloudProvider,
		"The provider for cloud services. Empty string for no provider.")

	fs.StringVar(&s.CloudConfigFile, "cloud-config", s.CloudConfigFile,
		"The path to the cloud provider configuration file. Empty string for no configuration file.")
}
