package cert

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/client-go/util/keyutil"
	netutils "k8s.io/utils/net"
)

const duration365d = time.Hour * 24 * 365

type Config struct {
	CommonName   string
	Organization []string
	AltNames     AltNames
	Usages       []x509.ExtKeyUsage
}

type AltNames struct {
	DNSNames []string
	IPs      []net.IP
}

func GenerateSelfSignedCertKey(host string, alternateIPs []net.IP, alternateDNS []string) ([]byte, []byte, error) {
	return GenerateSelfSignedCertKeyWithFixtures(host, alternateIPs, alternateDNS, "")
}

func GenerateSelfSignedCertKeyWithFixtures(host string, alternateIPs []net.IP, alternateDNS []string, fixtureDirectory string) ([]byte, []byte, error) {
	validFrom := time.Now().Add(-time.Hour)
	maxAge := time.Hour * 24 * 365

	baseName := fmt.Sprintf("%s_%s_%s", host, strings.Join(ipsToStrings(alternateIPs), "-"), strings.Join(alternateDNS, "-"))
	certFixturePath := filepath.Join(fixtureDirectory, baseName+".crt")
	keyFixturePath := filepath.Join(fixtureDirectory, baseName+".key")
	if len(fixtureDirectory) > 0 {
		// 已经存在就不要生成了
		cert, err := ioutil.ReadFile(certFixturePath)
		if err == nil {
			key, err := ioutil.ReadFile(keyFixturePath)
			if err == nil {
				return cert, key, nil
			}
			return nil, nil, fmt.Errorf("cert %s can be read, but key %s cannot: %v", certFixturePath, keyFixturePath, err)
		}
		maxAge = 100 * time.Hour * 24 * 365 // 100 years fixtures
	}

	// 产生2048位的rsa key
	caKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("%s-ca@%d", host, time.Now().Unix()),
		},
		NotBefore: validFrom,
		NotAfter:  validFrom.Add(maxAge),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// 传入证书和父证书模板，公私钥就可以生成自签名的CA证书字节数组
	caDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	// 将字节数组转为证书格式
	caCertificate, err := x509.ParseCertificate(caDERBytes)
	if err != nil {
		return nil, nil, err
	}

	priv, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("%s@%d", host, time.Now().Unix()),
		},
		NotBefore: validFrom,
		NotAfter:  validFrom.Add(maxAge),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 是ip就加到ip列表，是名字就加dns名字列表
	if ip := netutils.ParseIPSloppy(host); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, host)
	}

	template.IPAddresses = append(template.IPAddresses, alternateIPs...)
	template.DNSNames = append(template.DNSNames, alternateDNS...)

	// 用刚生成的自签名证书签发一个，序号为2
	derBytes, err := x509.CreateCertificate(cryptorand.Reader, &template, caCertificate, &priv.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	certBuffer := bytes.Buffer{}
	// 序号2的证书
	if err := pem.Encode(&certBuffer, &pem.Block{Type: CertificateBlockType, Bytes: derBytes}); err != nil {
		return nil, nil, err
	}
	// 序号1的证书
	if err := pem.Encode(&certBuffer, &pem.Block{Type: CertificateBlockType, Bytes: caDERBytes}); err != nil {
		return nil, nil, err
	}

	keyBuffer := bytes.Buffer{}
	if err := pem.Encode(&keyBuffer, &pem.Block{Type: keyutil.RSAPrivateKeyBlockType, Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return nil, nil, err
	}

	if len(fixtureDirectory) > 0 {
		if err := ioutil.WriteFile(certFixturePath, certBuffer.Bytes(), 0644); err != nil {
			return nil, nil, fmt.Errorf("failed to write cert fixture to %s: %v", certFixturePath, err)
		}
		if err := ioutil.WriteFile(keyFixturePath, keyBuffer.Bytes(), 0644); err != nil {
			return nil, nil, fmt.Errorf("failed to write key fixture to %s: %v", certFixturePath, err)
		}
	}
	// 返回的是2号证书和1号证书的结合体，还有私钥，1号证书的私钥不会保存
	return certBuffer.Bytes(), keyBuffer.Bytes(), nil
}

func ipsToStrings(ips []net.IP) []string {
	ss := make([]string, 0, len(ips))
	for _, ip := range ips {
		ss = append(ss, ip.String())
	}
	return ss
}
