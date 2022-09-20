package keyutil

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	// ECPrivateKeyBlockType is a possible value for pem.Block.Type.
	ECPrivateKeyBlockType = "EC PRIVATE KEY"
	// RSAPrivateKeyBlockType is a possible value for pem.Block.Type.
	RSAPrivateKeyBlockType = "RSA PRIVATE KEY"
	// PrivateKeyBlockType is a possible value for pem.Block.Type.
	PrivateKeyBlockType = "PRIVATE KEY"
	// PublicKeyBlockType is a possible value for pem.Block.Type.
	PublicKeyBlockType = "PUBLIC KEY"
)

func MakeEllipticPrivateKeyPEM() ([]byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), cryptorand.Reader)
	if err != nil {
		return nil, err
	}

	derBytes, err := x509.MarshalECPrivateKey((privateKey))
	if err != nil {
		return nil, err
	}

	privateKeyPemBlock := &pem.Block{
		Type:  ECPrivateKeyBlockType,
		Bytes: derBytes,
	}
	return pem.EncodeToMemory(privateKeyPemBlock), nil
}

func WriteKey(keyPath string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(keyPath), os.FileMode(0755)); err != nil {
		return err
	}
	return ioutil.WriteFile(keyPath, data, os.FileMode(0600))
}

func LoadOrGenerateKeyFile(keyPath string) (data []byte, wasGenerated bool, err error) {
	loadedData, err := ioutil.ReadFile(keyPath)
	// Call verifyKeyData to ensure the file wasn't empty/corrupt.
	if err == nil && verifyKeyData(loadedData) {
		return loadedData, false, err
	}
	if !os.IsNotExist(err) {
		return nil, false, fmt.Errorf("error loading key from %s: %v", keyPath, err)
	}

	generatedData, err := MakeEllipticPrivateKeyPEM()
	if err != nil {
		return nil, false, fmt.Errorf("error generating key: %v", err)
	}
	if err := WriteKey(keyPath, generatedData); err != nil {
		return nil, false, fmt.Errorf("error writing key to %s: %v", keyPath, err)
	}
	return generatedData, true, nil
}

func MarshalPrivateKeyToPEM(privateKey crypto.PrivateKey) ([]byte, error) {
	switch t := privateKey.(type) {
	case *ecdsa.PrivateKey:
		derBytes, err := x509.MarshalECPrivateKey(t)
		if err != nil {
			return nil, err
		}
		block := &pem.Block{
			Type:  ECPrivateKeyBlockType,
			Bytes: derBytes,
		}
		return pem.EncodeToMemory(block), nil
	case *rsa.PrivateKey:
		block := &pem.Block{
			Type:  RSAPrivateKeyBlockType,
			Bytes: x509.MarshalPKCS1PrivateKey(t),
		}
		return pem.EncodeToMemory(block), nil
	default:
		return nil, fmt.Errorf("private key is not a recognized type: %T", privateKey)
	}
}

func PrivateKeyFromFile(file string) (interface{}, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	key, err := ParsePrivateKeyPEM(data)
	if err != nil {
		return nil, fmt.Errorf("error reading private key file %s: %v", file, err)
	}
	return key, nil
}

func PublicKeysFromFile(file string) ([]interface{}, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	keys, err := ParsePublicKeysPEM(data)
	if err != nil {
		return nil, fmt.Errorf("error reading public key file %s: %v", file, err)
	}
	return keys, nil
}

func verifyKeyData(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	_, err := ParsePrivateKeyPEM(data)
	return err == nil
}

func ParsePrivateKeyPEM(keyData []byte) (interface{}, error) {
	var privateKeyPemBlock *pem.Block
	for {
		privateKeyPemBlock, keyData = pem.Decode(keyData)
		if privateKeyPemBlock == nil {
			break
		}
		switch privateKeyPemBlock.Type {
		case ECPrivateKeyBlockType:
			if key, err := x509.ParseECPrivateKey(privateKeyPemBlock.Bytes); err == nil {
				return key, nil
			}
		case RSAPrivateKeyBlockType:
			// RSA Private Key in PKCS#1 format
			if key, err := x509.ParsePKCS1PrivateKey(privateKeyPemBlock.Bytes); err == nil {
				return key, nil
			}
		case PrivateKeyBlockType:
			// RSA or ECDSA Private Key in unencrypted PKCS#8 format
			if key, err := x509.ParsePKCS8PrivateKey(privateKeyPemBlock.Bytes); err == nil {
				return key, nil
			}
		}
	}
	return nil, fmt.Errorf("data does not contain a valid RSA or ECDSA private key")
}

func ParsePublicKeysPEM(keyData []byte) ([]interface{}, error) {
	var block *pem.Block
	keys := []interface{}{}
	for {
		block, keyData = pem.Decode(keyData)
		if block == nil {
			break
		}

		if privateKey, err := parseRSAPrivateKey(block.Bytes); err == nil {
			keys = append(keys, &privateKey.PublicKey)
			continue
		}
		if publicKey, err := parseRSAPublicKey(block.Bytes); err == nil {
			keys = append(keys, publicKey)
			continue
		}
		if privateKey, err := parseECPrivateKey(block.Bytes); err == nil {
			keys = append(keys, &privateKey.PublicKey)
			continue
		}
		if publicKey, err := parseECPublicKey(block.Bytes); err == nil {
			keys = append(keys, publicKey)
			continue
		}
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("data does not contain any valid RSA or ECDSA public keys")
	}
	return keys, nil
}

func parseRSAPublicKey(data []byte) (*rsa.PublicKey, error) {
	var err error

	// Parse the key
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKIXPublicKey(data); err != nil {
		if cert, err := x509.ParseCertificate(data); err == nil {
			parsedKey = cert.PublicKey
		} else {
			return nil, err
		}
	}
	var pubKey *rsa.PublicKey
	var ok bool

	if pubKey, ok = parsedKey.(*rsa.PublicKey); !ok {
		return nil, fmt.Errorf("data doesn't contain valid RSA Public Key")
	}
	return pubKey, nil
}

func parseRSAPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	var err error

	// Parse the key
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(data); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(data); err != nil {
			return nil, err
		}
	}

	// Test if parsed key is an RSA Private Key
	var privKey *rsa.PrivateKey
	var ok bool
	if privKey, ok = parsedKey.(*rsa.PrivateKey); !ok {
		return nil, fmt.Errorf("data doesn't contain valid RSA Private Key")
	}

	return privKey, nil
}

func parseECPublicKey(data []byte) (*ecdsa.PublicKey, error) {
	var err error

	// Parse the key
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKIXPublicKey(data); err != nil {
		if cert, err := x509.ParseCertificate(data); err == nil {
			parsedKey = cert.PublicKey
		} else {
			return nil, err
		}
	}

	// Test if parsed key is an ECDSA Public Key
	var pubKey *ecdsa.PublicKey
	var ok bool
	if pubKey, ok = parsedKey.(*ecdsa.PublicKey); !ok {
		return nil, fmt.Errorf("data doesn't contain valid ECDSA Public Key")
	}

	return pubKey, nil
}

func parseECPrivateKey(data []byte) (*ecdsa.PrivateKey, error) {
	var err error

	// Parse the key
	var parsedKey interface{}
	if parsedKey, err = x509.ParseECPrivateKey(data); err != nil {
		return nil, err
	}

	// Test if parsed key is an ECDSA Private Key
	var privKey *ecdsa.PrivateKey
	var ok bool
	if privKey, ok = parsedKey.(*ecdsa.PrivateKey); !ok {
		return nil, fmt.Errorf("data doesn't contain valid ECDSA Private Key")
	}

	return privKey, nil
}
