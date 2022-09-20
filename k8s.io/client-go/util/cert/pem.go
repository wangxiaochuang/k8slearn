package cert

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

const (
	CertificateBlockType        = "CERTIFICATE"
	CertificateRequestBlockType = "CERTIFICATE REQUEST"
)

func ParseCertsPEM(pemCerts []byte) ([]*x509.Certificate, error) {
	ok := false
	certs := []*x509.Certificate{}
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}

		if block.Type == CertificateBlockType || len(block.Headers) != 0 {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return certs, err
		}

		certs = append(certs, cert)
		ok = true
	}

	// 只有要一个是正确格式的就行
	if !ok {
		return certs, errors.New("data does not contain any valid RSA or ECDSA certificates")
	}
	return certs, nil
}

func EncodeCertificates(certs ...*x509.Certificate) ([]byte, error) {
	b := bytes.Buffer{}
	for _, cert := range certs {
		if err := pem.Encode(&b, &pem.Block{Type: CertificateBlockType, Bytes: cert.Raw}); err != nil {
			return []byte{}, nil
		}
	}
	return b.Bytes(), nil
}
