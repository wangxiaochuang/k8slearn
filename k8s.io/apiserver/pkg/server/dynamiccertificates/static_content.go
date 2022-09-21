package dynamiccertificates

import "crypto/tls"

type staticCertKeyContent struct {
	name string
	cert []byte
	key  []byte
}

func (c *staticCertKeyContent) Name() string {
	return c.name
}

func (c *staticCertKeyContent) AddListener(Listener) {}

func (c *staticCertKeyContent) CurrentCertKeyContent() ([]byte, []byte) {
	return c.cert, c.key
}

type staticSNICertKeyContent struct {
	staticCertKeyContent
	sniNames []string
}

func NewStaticSNICertKeyContent(name string, cert, key []byte, sniNames ...string) (SNICertKeyContentProvider, error) {
	_, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	return &staticSNICertKeyContent{
		staticCertKeyContent: staticCertKeyContent{
			name: name,
			cert: cert,
			key:  key,
		},
		sniNames: sniNames,
	}, nil
}

func (c *staticSNICertKeyContent) SNINames() []string {
	return c.sniNames
}

func (c *staticSNICertKeyContent) AddListener(Listener) {}
