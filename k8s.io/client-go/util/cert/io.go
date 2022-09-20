package cert

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func CanReadCertAndKey(certPath, keyPath string) (bool, error) {
	certReadable := canReadFile(certPath)
	keyReadable := canReadFile(keyPath)

	if certReadable == false && keyReadable == false {
		return false, nil
	}

	if certReadable == false {
		return false, fmt.Errorf("error reading %s, certificate and key must be supplied as a pair", certPath)
	}

	if keyReadable == false {
		return false, fmt.Errorf("error reading %s, certificate and key must be supplied as a pair", keyPath)
	}

	return true, nil
}

func canReadFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}

	defer f.Close()

	return true
}

func WriteCert(certPath string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(certPath), os.FileMode(0755)); err != nil {
		return err
	}
	return ioutil.WriteFile(certPath, data, os.FileMode(0644))
}

func NewPool(filename string) (*x509.CertPool, error) {
	pemBlock, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	pool, err := NewPoolFromBytes(pemBlock)
	if err != nil {
		return nil, fmt.Errorf("error creating pool from %s: %s", filename, err)
	}
	return pool, nil
}

func NewPoolFromBytes(pemBlock []byte) (*x509.CertPool, error) {
	certs, err := ParseCertsPEM(pemBlock)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	for _, cert := range certs {
		pool.AddCert(cert)
	}
	return pool, nil
}

func CertsFromFile(file string) ([]*x509.Certificate, error) {
	pemBlock, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	certs, err := ParseCertsPEM(pemBlock)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %s", file, err)
	}
	return certs, nil
}
