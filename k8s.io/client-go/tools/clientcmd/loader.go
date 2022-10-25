package clientcmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	goruntime "runtime"
	"strings"

	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	"k8s.io/client-go/util/homedir"
)

const (
	RecommendedConfigPathFlag   = "kubeconfig"
	RecommendedConfigPathEnvVar = "KUBECONFIG"
	RecommendedHomeDir          = ".kube"
	RecommendedFileName         = "config"
	RecommendedSchemaName       = "schema"
)

var (
	RecommendedConfigDir  = filepath.Join(homedir.HomeDir(), RecommendedHomeDir)
	RecommendedHomeFile   = filepath.Join(RecommendedConfigDir, RecommendedFileName)
	RecommendedSchemaFile = filepath.Join(RecommendedConfigDir, RecommendedSchemaName)
)

func currentMigrationRules() map[string]string {
	var oldRecommendedHomeFileName string
	if goruntime.GOOS == "windows" {
		oldRecommendedHomeFileName = RecommendedFileName
	} else {
		oldRecommendedHomeFileName = ".kubeconfig"
	}
	return map[string]string{
		RecommendedHomeFile: filepath.Join(os.Getenv("HOME"), RecommendedHomeDir, oldRecommendedHomeFileName),
	}
}

type ClientConfigLoader interface {
	ConfigAccess
	// IsDefaultConfig returns true if the returned config matches the defaults.
	IsDefaultConfig(*restclient.Config) bool
	// Load returns the latest config
	Load() (*clientcmdapi.Config, error)
}

type KubeconfigGetter func() (*clientcmdapi.Config, error)

type ClientConfigGetter struct {
	kubeconfigGetter KubeconfigGetter
}

var _ ClientConfigLoader = &ClientConfigGetter{}

func (g *ClientConfigGetter) Load() (*clientcmdapi.Config, error) {
	return g.kubeconfigGetter()
}

func (g *ClientConfigGetter) GetLoadingPrecedence() []string {
	return nil
}
func (g *ClientConfigGetter) GetStartingConfig() (*clientcmdapi.Config, error) {
	return g.kubeconfigGetter()
}
func (g *ClientConfigGetter) GetDefaultFilename() string {
	return ""
}
func (g *ClientConfigGetter) IsExplicitFile() bool {
	return false
}
func (g *ClientConfigGetter) GetExplicitFile() string {
	return ""
}
func (g *ClientConfigGetter) IsDefaultConfig(config *restclient.Config) bool {
	return false
}

type ClientConfigLoadingRules struct {
	ExplicitPath string
	Precedence   []string

	// MigrationRules is a map of destination files to source files.  If a destination file is not present, then the source file is checked.
	// If the source file is present, then it is copied to the destination file BEFORE any further loading happens.
	MigrationRules map[string]string

	// DoNotResolvePaths indicates whether or not to resolve paths with respect to the originating files.  This is phrased as a negative so
	// that a default object that doesn't set this will usually get the behavior it wants.
	DoNotResolvePaths bool

	// DefaultClientConfig is an optional field indicating what rules to use to calculate a default configuration.
	// This should match the overrides passed in to ClientConfig loader.
	DefaultClientConfig ClientConfig

	// WarnIfAllMissing indicates whether the configuration files pointed by KUBECONFIG environment variable are present or not.
	// In case of missing files, it warns the user about the missing files.
	WarnIfAllMissing bool
}

// ClientConfigLoadingRules implements the ClientConfigLoader interface.
var _ ClientConfigLoader = &ClientConfigLoadingRules{}

func NewDefaultClientConfigLoadingRules() *ClientConfigLoadingRules {
	chain := []string{}
	warnIfAllMissing := false

	envVarFiles := os.Getenv(RecommendedConfigPathEnvVar)
	if len(envVarFiles) != 0 {
		fileList := filepath.SplitList(envVarFiles)
		// prevent the same path load multiple times
		chain = append(chain, deduplicate(fileList)...)
		warnIfAllMissing = true

	} else {
		chain = append(chain, RecommendedHomeFile)
	}

	return &ClientConfigLoadingRules{
		Precedence:       chain,
		MigrationRules:   currentMigrationRules(),
		WarnIfAllMissing: warnIfAllMissing,
	}
}

func (rules *ClientConfigLoadingRules) Load() (*clientcmdapi.Config, error) {
	panic("not implemented")
}

func (rules *ClientConfigLoadingRules) Migrate() error {
	panic("not implemented")
}

func (rules *ClientConfigLoadingRules) GetLoadingPrecedence() []string {
	if len(rules.ExplicitPath) > 0 {
		return []string{rules.ExplicitPath}
	}

	return rules.Precedence
}

func (rules *ClientConfigLoadingRules) GetStartingConfig() (*clientcmdapi.Config, error) {
	panic("not implemented")
}

func (rules *ClientConfigLoadingRules) GetDefaultFilename() string {
	// Explicit file if we have one.
	if rules.IsExplicitFile() {
		return rules.GetExplicitFile()
	}
	// Otherwise, first existing file from precedence.
	for _, filename := range rules.GetLoadingPrecedence() {
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}
	// If none exists, use the first from precedence.
	if len(rules.Precedence) > 0 {
		return rules.Precedence[0]
	}
	return ""
}

func (rules *ClientConfigLoadingRules) IsExplicitFile() bool {
	return len(rules.ExplicitPath) > 0
}

// GetExplicitFile implements ConfigAccess
func (rules *ClientConfigLoadingRules) GetExplicitFile() string {
	return rules.ExplicitPath
}

// IsDefaultConfig returns true if the provided configuration matches the default
func (rules *ClientConfigLoadingRules) IsDefaultConfig(config *restclient.Config) bool {
	if rules.DefaultClientConfig == nil {
		return false
	}
	defaultConfig, err := rules.DefaultClientConfig.ClientConfig()
	if err != nil {
		return false
	}
	return reflect.DeepEqual(config, defaultConfig)
}

func LoadFromFile(filename string) (*clientcmdapi.Config, error) {
	kubeconfigBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	config, err := Load(kubeconfigBytes)
	if err != nil {
		return nil, err
	}
	klog.V(6).Infoln("Config loaded from file: ", filename)

	// set LocationOfOrigin on every Cluster, User, and Context
	for key, obj := range config.AuthInfos {
		obj.LocationOfOrigin = filename
		config.AuthInfos[key] = obj
	}
	for key, obj := range config.Clusters {
		obj.LocationOfOrigin = filename
		config.Clusters[key] = obj
	}
	for key, obj := range config.Contexts {
		obj.LocationOfOrigin = filename
		config.Contexts[key] = obj
	}

	if config.AuthInfos == nil {
		config.AuthInfos = map[string]*clientcmdapi.AuthInfo{}
	}
	if config.Clusters == nil {
		config.Clusters = map[string]*clientcmdapi.Cluster{}
	}
	if config.Contexts == nil {
		config.Contexts = map[string]*clientcmdapi.Context{}
	}

	return config, nil
}

func Load(data []byte) (*clientcmdapi.Config, error) {
	config := clientcmdapi.NewConfig()
	// if there's no data in a file, return the default object instead of failing (DecodeInto reject empty input)
	if len(data) == 0 {
		return config, nil
	}
	decoded, _, err := clientcmdlatest.Codec.Decode(data, &schema.GroupVersionKind{Version: clientcmdlatest.Version, Kind: "Config"}, config)
	if err != nil {
		return nil, err
	}
	return decoded.(*clientcmdapi.Config), nil
}

func WriteToFile(config clientcmdapi.Config, filename string) error {
	content, err := Write(config)
	if err != nil {
		return err
	}
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(filename, content, 0600); err != nil {
		return err
	}
	return nil
}

func lockFile(filename string) error {
	// TODO: find a way to do this with actual file locks. Will
	// probably need separate solution for windows and Linux.

	// Make sure the dir exists before we try to create a lock file.
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(lockName(filename), os.O_CREATE|os.O_EXCL, 0)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func unlockFile(filename string) error {
	return os.Remove(lockName(filename))
}

func lockName(filename string) string {
	return filename + ".lock"
}

// Write serializes the config to yaml.
// Encapsulates serialization without assuming the destination is a file.
func Write(config clientcmdapi.Config) ([]byte, error) {
	return runtime.Encode(clientcmdlatest.Codec, &config)
}

func (rules ClientConfigLoadingRules) ResolvePaths() bool {
	return !rules.DoNotResolvePaths
}

func ResolveLocalPaths(config *clientcmdapi.Config) error {
	panic("not implemented")
}

func RelativizeClusterLocalPaths(cluster *clientcmdapi.Cluster) error {
	panic("not implemented")
}

func RelativizeAuthInfoLocalPaths(authInfo *clientcmdapi.AuthInfo) error {
	panic("not implemented")
}

func RelativizeConfigPaths(config *clientcmdapi.Config, base string) error {
	return RelativizePathWithNoBacksteps(GetConfigFileReferences(config), base)
}

func ResolveConfigPaths(config *clientcmdapi.Config, base string) error {
	return ResolvePaths(GetConfigFileReferences(config), base)
}

func GetConfigFileReferences(config *clientcmdapi.Config) []*string {
	panic("not implemented")
}

func GetAuthInfoFileReferences(authInfo *clientcmdapi.AuthInfo) []*string {
	s := []*string{&authInfo.ClientCertificate, &authInfo.ClientKey, &authInfo.TokenFile}
	// Only resolve exec command if it isn't PATH based.
	if authInfo.Exec != nil && strings.ContainsRune(authInfo.Exec.Command, filepath.Separator) {
		s = append(s, &authInfo.Exec.Command)
	}
	return s
}

func ResolvePaths(refs []*string, base string) error {
	for _, ref := range refs {
		// Don't resolve empty paths
		if len(*ref) > 0 {
			// Don't resolve absolute paths
			if !filepath.IsAbs(*ref) {
				*ref = filepath.Join(base, *ref)
			}
		}
	}
	return nil
}

func RelativizePathWithNoBacksteps(refs []*string, base string) error {
	panic("not implemented")
}

func MakeRelative(path, base string) (string, error) {
	if len(path) > 0 {
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return path, err
		}
		return rel, nil
	}
	return path, nil
}

func deduplicate(s []string) []string {
	encountered := map[string]bool{}
	ret := make([]string, 0)
	for i := range s {
		if encountered[s[i]] {
			continue
		}
		encountered[s[i]] = true
		ret = append(ret, s[i])
	}
	return ret
}
