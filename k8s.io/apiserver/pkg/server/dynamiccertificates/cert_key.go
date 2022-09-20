package dynamiccertificates

import "bytes"

type certKeyContent struct {
	cert []byte
	key  []byte
}

func (c *certKeyContent) Equal(rhs *certKeyContent) bool {
	if c == nil || rhs == nil {
		return c == rhs
	}

	return bytes.Equal(c.key, rhs.key) && bytes.Equal(c.cert, rhs.cert)
}

type sniCertKeyContent struct {
	certKeyContent
	sniNames []string
}

func (c *sniCertKeyContent) Equal(rhs *sniCertKeyContent) bool {
	if c == nil || rhs == nil {
		return c == rhs
	}

	if len(c.sniNames) != len(rhs.sniNames) {
		return false
	}

	for i := range c.sniNames {
		if c.sniNames[i] != rhs.sniNames[i] {
			return false
		}
	}

	return c.certKeyContent.Equal(&rhs.certKeyContent)
}
