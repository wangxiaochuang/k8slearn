package flag

import (
	"crypto/tls"
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	ciphers         = map[string]uint16{}
	insecureCiphers = map[string]uint16{}
)

func init() {
	for _, suite := range tls.CipherSuites() {
		ciphers[suite.Name] = suite.ID
	}
	// keep legacy names for backward compatibility
	ciphers["TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305"] = tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
	ciphers["TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305"] = tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256

	for _, suite := range tls.InsecureCipherSuites() {
		insecureCiphers[suite.Name] = suite.ID
	}
}

func InsecureTLSCiphers() map[string]uint16 {
	cipherKeys := make(map[string]uint16, len(insecureCiphers))
	for k, v := range insecureCiphers {
		cipherKeys[k] = v
	}
	return cipherKeys
}

func InsecureTLSCipherNames() []string {
	cipherKeys := sets.NewString()
	for key := range insecureCiphers {
		cipherKeys.Insert(key)
	}
	return cipherKeys.List()
}

func PreferredTLSCipherNames() []string {
	cipherKeys := sets.NewString()
	for key := range ciphers {
		cipherKeys.Insert(key)
	}
	return cipherKeys.List()
}

func allCiphers() map[string]uint16 {
	acceptedCiphers := make(map[string]uint16, len(ciphers)+len(insecureCiphers))
	for k, v := range ciphers {
		acceptedCiphers[k] = v
	}
	for k, v := range insecureCiphers {
		acceptedCiphers[k] = v
	}
	return acceptedCiphers
}

func TLSCipherPossibleValues() []string {
	cipherKeys := sets.NewString()
	acceptedCiphers := allCiphers()
	for key := range acceptedCiphers {
		cipherKeys.Insert(key)
	}
	return cipherKeys.List()
}

func TLSCipherSuites(cipherNames []string) ([]uint16, error) {
	if len(cipherNames) == 0 {
		return nil, nil
	}
	ciphersIntSlice := make([]uint16, 0)
	possibleCiphers := allCiphers()
	for _, cipher := range cipherNames {
		intValue, ok := possibleCiphers[cipher]
		if !ok {
			return nil, fmt.Errorf("Cipher suite %s not supported or doesn't exist", cipher)
		}
		ciphersIntSlice = append(ciphersIntSlice, intValue)
	}
	return ciphersIntSlice, nil
}

var versions = map[string]uint16{
	"VersionTLS10": tls.VersionTLS10,
	"VersionTLS11": tls.VersionTLS11,
	"VersionTLS12": tls.VersionTLS12,
	"VersionTLS13": tls.VersionTLS13,
}

func TLSPossibleVersions() []string {
	versionsKeys := sets.NewString()
	for key := range versions {
		versionsKeys.Insert(key)
	}
	return versionsKeys.List()
}

func TLSVersion(versionName string) (uint16, error) {
	if len(versionName) == 0 {
		return DefaultTLSVersion(), nil
	}
	if version, ok := versions[versionName]; ok {
		return version, nil
	}
	return 0, fmt.Errorf("unknown tls version %q", versionName)
}

func DefaultTLSVersion() uint16 {
	return tls.VersionTLS12
}
