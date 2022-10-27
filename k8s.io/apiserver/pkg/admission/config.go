package admission

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/apis/apiserver"
	apiserverv1 "k8s.io/apiserver/pkg/apis/apiserver/v1"
)

func makeAbs(path, base string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	if len(base) == 0 || base == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		base = cwd
	}
	return filepath.Join(base, path), nil
}

func ReadAdmissionConfiguration(pluginNames []string, configFilePath string, configScheme *runtime.Scheme) (ConfigProvider, error) {
	if configFilePath == "" {
		return configProvider{config: &apiserver.AdmissionConfiguration{}}, nil
	}

	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read admission control configuration from %q [%v]", configFilePath, err)
	}
	codecs := serializer.NewCodecFactory(configScheme)
	decoder := codecs.UniversalDecoder()
	decodedObj, err := runtime.Decode(decoder, data)

	if err == nil {
		decodedConfig, ok := decodedObj.(*apiserver.AdmissionConfiguration)
		if !ok {
			return nil, fmt.Errorf("unexpected type: %T", decodedObj)
		}
		baseDir := path.Dir(configFilePath)
		for i := range decodedConfig.Plugins {
			if decodedConfig.Plugins[i].Path == "" {
				continue
			}
			// we update relative file paths to absolute paths
			absPath, err := makeAbs(decodedConfig.Plugins[i].Path, baseDir)
			if err != nil {
				return nil, err
			}
			decodedConfig.Plugins[i].Path = absPath
		}
		return configProvider{
			config: decodedConfig,
		}, nil
	}

	if !(runtime.IsMissingVersion(err) || runtime.IsMissingKind(err) || runtime.IsNotRegisteredError(err)) {
		return nil, err
	}

	unstructuredData := map[string]interface{}{}
	if err2 := yaml.Unmarshal(data, &unstructuredData); err2 != nil {
		return nil, err
	}
	_, isLegacyImagePolicy := unstructuredData["imagePolicy"]
	_, isLegacyPodNodeSelector := unstructuredData["podNodeSelectorPluginConfig"]
	if !isLegacyImagePolicy && !isLegacyPodNodeSelector {
		return nil, err
	}

	legacyPluginsWithUnversionedConfig := sets.NewString("ImagePolicyWebhook", "PodNodeSelector")
	externalConfig := &apiserverv1.AdmissionConfiguration{}
	for _, pluginName := range pluginNames {
		if legacyPluginsWithUnversionedConfig.Has(pluginName) {
			externalConfig.Plugins = append(externalConfig.Plugins,
				apiserverv1.AdmissionPluginConfiguration{
					Name: pluginName,
					Path: configFilePath})
		}
	}
	configScheme.Default(externalConfig)
	internalConfig := &apiserver.AdmissionConfiguration{}
	if err := configScheme.Convert(externalConfig, internalConfig, nil); err != nil {
		return nil, err
	}
	return configProvider{
		config: internalConfig,
	}, nil
}

type configProvider struct {
	config *apiserver.AdmissionConfiguration
}

func GetAdmissionPluginConfigurationFor(pluginCfg apiserver.AdmissionPluginConfiguration) (io.Reader, error) {
	// if there is a nest object, return it directly
	if pluginCfg.Configuration != nil {
		return bytes.NewBuffer(pluginCfg.Configuration.Raw), nil
	}
	// there is nothing nested, so we delegate to path
	if pluginCfg.Path != "" {
		content, err := ioutil.ReadFile(pluginCfg.Path)
		if err != nil {
			klog.Fatalf("Couldn't open admission plugin configuration %s: %#v", pluginCfg.Path, err)
			return nil, err
		}
		return bytes.NewBuffer(content), nil
	}
	// there is no special config at all
	return nil, nil
}

func (p configProvider) ConfigFor(pluginName string) (io.Reader, error) {
	// there is no config, so there is no potential config
	if p.config == nil {
		return nil, nil
	}
	// look for matching plugin and get configuration
	for _, pluginCfg := range p.config.Plugins {
		if pluginName != pluginCfg.Name {
			continue
		}
		pluginConfig, err := GetAdmissionPluginConfigurationFor(pluginCfg)
		if err != nil {
			return nil, err
		}
		return pluginConfig, nil
	}
	// there is no registered config that matches on plugin name.
	return nil, nil
}
