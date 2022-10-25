package clientcmd

import clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

type ConfigAccess interface {
	// GetLoadingPrecedence returns the slice of files that should be used for loading and inspecting the config
	GetLoadingPrecedence() []string
	// GetStartingConfig returns the config that subcommands should being operating against.  It may or may not be merged depending on loading rules
	GetStartingConfig() (*clientcmdapi.Config, error)
	// GetDefaultFilename returns the name of the file you should write into (create if necessary), if you're trying to create a new stanza as opposed to updating an existing one.
	GetDefaultFilename() string
	// IsExplicitFile indicates whether or not this command is interested in exactly one file.  This implementation only ever does that  via a flag, but implementations that handle local, global, and flags may have more
	IsExplicitFile() bool
	// GetExplicitFile returns the particular file this command is operating against.  This implementation only ever has one, but implementations that handle local, global, and flags may have more
	GetExplicitFile() string
}

type PathOptions struct {
	// GlobalFile is the full path to the file to load as the global (final) option
	GlobalFile string
	// EnvVar is the env var name that points to the list of kubeconfig files to load
	EnvVar string
	// ExplicitFileFlag is the name of the flag to use for prompting for the kubeconfig file
	ExplicitFileFlag string

	// GlobalFileSubpath is an optional value used for displaying help
	GlobalFileSubpath string

	LoadingRules *ClientConfigLoadingRules
}

var (
	UseModifyConfigLock = true
)
