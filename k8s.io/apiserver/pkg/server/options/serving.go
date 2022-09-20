package options

import (
	"context"
	"fmt"
	"net"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/pflag"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	netutils "k8s.io/utils/net"
)

type SecureServingOptions struct {
	BindAddress                  net.IP
	BindPort                     int
	BindNetwork                  string
	Required                     bool
	ExternalAddress              net.IP
	Listener                     net.Listener
	ServerCert                   GeneratableKeyCert
	SNICertKeys                  []cliflag.NamedCertKey
	CipherSuites                 []string
	MinTLSVersion                string
	HTTP2MaxStreamsPerConnection int

	PermitPortSharing bool

	PermitAddressSharing bool
}

type CertKey struct {
	CertFile string
	KeyFile  string
}

type GeneratableKeyCert struct {
	CertKey       CertKey
	CertDirectory string
	PairName      string

	GeneratedCert    dynamiccertificates.CertKeyContentProvider
	FixtureDirectory string
}

func NewSecureServingOptions() *SecureServingOptions {
	return &SecureServingOptions{
		BindAddress: netutils.ParseIPSloppy("0.0.0.0"),
		BindPort:    443,
		ServerCert: GeneratableKeyCert{
			PairName:      "apiserver",
			CertDirectory: "apiserver.local.config/certificates",
		},
	}
}

func (s *SecureServingOptions) DefaultExternalAddress() (net.IP, error) {
	if s.ExternalAddress != nil && !s.ExternalAddress.IsUnspecified() {
		return s.ExternalAddress, nil
	}
	return utilnet.ResolveBindAddress(s.BindAddress)
}

func (s *SecureServingOptions) Validate() []error {
	if s == nil {
		return nil
	}

	errors := []error{}

	if s.Required && s.BindPort < 1 || s.BindPort > 65535 {
		errors = append(errors, fmt.Errorf("--secure-port %v must be between 1 and 65535, inclusive. It cannot be turned off with 0", s.BindPort))
	} else if s.BindPort < 0 || s.BindPort > 65535 {
		errors = append(errors, fmt.Errorf("--secure-port %v must be between 0 and 65535, inclusive. 0 for turning off secure port", s.BindPort))
	}

	if (len(s.ServerCert.CertKey.CertFile) != 0 || len(s.ServerCert.CertKey.KeyFile) != 0) && s.ServerCert.GeneratedCert != nil {
		errors = append(errors, fmt.Errorf("cert/key file and in-memory certificate cannot both be set"))
	}

	return errors
}

// p148
func (s *SecureServingOptions) AddFlags(fs *pflag.FlagSet) {
	if s == nil {
		return
	}

	fs.IPVar(&s.BindAddress, "bind-address", s.BindAddress, ""+
		"The IP address on which to listen for the --secure-port port. The "+
		"associated interface(s) must be reachable by the rest of the cluster, and by CLI/web "+
		"clients. If blank or an unspecified address (0.0.0.0 or ::), all interfaces will be used.")

	desc := "The port on which to serve HTTPS with authentication and authorization."
	if s.Required {
		desc += " It cannot be switched off with 0."
	} else {
		desc += " If 0, don't serve HTTPS at all."
	}
	fs.IntVar(&s.BindPort, "secure-port", s.BindPort, desc)

	fs.StringVar(&s.ServerCert.CertDirectory, "cert-dir", s.ServerCert.CertDirectory, ""+
		"The directory where the TLS certs are located. "+
		"If --tls-cert-file and --tls-private-key-file are provided, this flag will be ignored.")

	fs.StringVar(&s.ServerCert.CertKey.CertFile, "tls-cert-file", s.ServerCert.CertKey.CertFile, ""+
		"File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated "+
		"after server cert). If HTTPS serving is enabled, and --tls-cert-file and "+
		"--tls-private-key-file are not provided, a self-signed certificate and key "+
		"are generated for the public address and saved to the directory specified by --cert-dir.")

	fs.StringVar(&s.ServerCert.CertKey.KeyFile, "tls-private-key-file", s.ServerCert.CertKey.KeyFile,
		"File containing the default x509 private key matching --tls-cert-file.")

	tlsCipherPreferredValues := cliflag.PreferredTLSCipherNames()
	tlsCipherInsecureValues := cliflag.InsecureTLSCipherNames()
	fs.StringSliceVar(&s.CipherSuites, "tls-cipher-suites", s.CipherSuites,
		"Comma-separated list of cipher suites for the server. "+
			"If omitted, the default Go cipher suites will be used. \n"+
			"Preferred values: "+strings.Join(tlsCipherPreferredValues, ", ")+". \n"+
			"Insecure values: "+strings.Join(tlsCipherInsecureValues, ", ")+".")

	tlsPossibleVersions := cliflag.TLSPossibleVersions()
	fs.StringVar(&s.MinTLSVersion, "tls-min-version", s.MinTLSVersion,
		"Minimum TLS version supported. "+
			"Possible values: "+strings.Join(tlsPossibleVersions, ", "))

	fs.Var(cliflag.NewNamedCertKeyArray(&s.SNICertKeys), "tls-sni-cert-key", ""+
		"A pair of x509 certificate and private key file paths, optionally suffixed with a list of "+
		"domain patterns which are fully qualified domain names, possibly with prefixed wildcard "+
		"segments. The domain patterns also allow IP addresses, but IPs should only be used if "+
		"the apiserver has visibility to the IP address requested by a client. "+
		"If no domain patterns are provided, the names of the certificate are "+
		"extracted. Non-wildcard matches trump over wildcard matches, explicit domain patterns "+
		"trump over extracted names. For multiple key/certificate pairs, use the "+
		"--tls-sni-cert-key multiple times. "+
		"Examples: \"example.crt,example.key\" or \"foo.crt,foo.key:*.foo.com,foo.com\".")

	fs.IntVar(&s.HTTP2MaxStreamsPerConnection, "http2-max-streams-per-connection", s.HTTP2MaxStreamsPerConnection, ""+
		"The limit that the server gives to clients for "+
		"the maximum number of streams in an HTTP/2 connection. "+
		"Zero means to use golang's default.")

	fs.BoolVar(&s.PermitPortSharing, "permit-port-sharing", s.PermitPortSharing,
		"If true, SO_REUSEPORT will be used when binding the port, which allows "+
			"more than one instance to bind on the same address and port. [default=false]")

	fs.BoolVar(&s.PermitAddressSharing, "permit-address-sharing", s.PermitAddressSharing,
		"If true, SO_REUSEADDR will be used when binding the port. This allows binding "+
			"to wildcard IPs like 0.0.0.0 and specific IPs in parallel, and it avoids waiting "+
			"for the kernel to release sockets in TIME_WAIT state. [default=false]")
}

func (s *SecureServingOptions) ApplyTo(config **server.SecureServingInfo) error {
	if s == nil {
		return nil
	}
	if s.BindPort <= 0 && s.Listener == nil {
		return nil
	}

	if s.Listener == nil {
		var err error
		addr := net.JoinHostPort(s.BindAddress.String(), strconv.Itoa(s.BindPort))

		c := net.ListenConfig{}

		ctls := multipleControls{}
		if s.PermitPortSharing {
			ctls = append(ctls, permitPortReuse)
		}
		if s.PermitAddressSharing {
			ctls = append(ctls, permitAddressReuse)
		}
		if len(ctls) > 0 {
			c.Control = ctls.Control
		}

		s.Listener, s.BindPort, err = CreateListener(s.BindNetwork, addr, c)
		if err != nil {
			return fmt.Errorf("failed to create listener: %v", err)
		}
	} else {
		if _, ok := s.Listener.Addr().(*net.TCPAddr); !ok {
			return fmt.Errorf("failed to parse ip and port from listener")
		}
		s.BindPort = s.Listener.Addr().(*net.TCPAddr).Port
		s.BindAddress = s.Listener.Addr().(*net.TCPAddr).IP
	}

	// 设置新的SecureServingInfo
	*config = &server.SecureServingInfo{
		Listener:                     s.Listener,
		HTTP2MaxStreamsPerConnection: s.HTTP2MaxStreamsPerConnection,
	}
	c := *config

	serverCertFile, serverKeyFile := s.ServerCert.CertKey.CertFile, s.ServerCert.CertKey.KeyFile
	if len(serverCertFile) != 0 || len(serverKeyFile) != 0 {
		var err error
		c.Cert, err = dynamiccertificates.NewDynamicServingContentFromFiles("serving-cert", serverCertFile, serverKeyFile)
		if err != nil {
			return err
		}
	} else if s.ServerCert.GeneratedCert != nil {
		c.Cert = s.ServerCert.GeneratedCert
	}

	if len(s.CipherSuites) != 0 {
		cipherSuites, err := cliflag.TLSCipherSuites(s.CipherSuites)
		if err != nil {
			return err
		}
		c.CipherSuites = cipherSuites
	}

	var err error
	c.MinTLSVersion, err = cliflag.TLSVersion(s.MinTLSVersion)
	if err != nil {
		return err
	}

	namedTLSCerts := make([]dynamiccertificates.SNICertKeyContentProvider, 0, len(s.SNICertKeys))
	for _, nck := range s.SNICertKeys {
		tlsCert, err := dynamiccertificates.NewDynamicSNIContentFromFiles("sni-serving-cert", nck.CertFile, nck.KeyFile, nck.Names...)
		namedTLSCerts = append(namedTLSCerts, tlsCert)
		if err != nil {
			return fmt.Errorf("failed to load SNI cert and key: %v", err)
		}
	}
	c.SNICerts = namedTLSCerts

	return nil
}

func (s *SecureServingOptions) MaybeDefaultWithSelfSignedCerts(publicAddress string, alternateDNS []string, alternateIPs []net.IP) error {
	if s == nil || (s.BindPort == 0 && s.Listener == nil) {
		return nil
	}
	keyCert := &s.ServerCert.CertKey
	// 如果指定了CertFile和KeyFile直接返回
	if len(keyCert.CertFile) != 0 || len(keyCert.KeyFile) != 0 {
		return nil
	}

	// 没有配置证书才会走到下面
	canReadCertAndKey := false
	if len(s.ServerCert.CertDirectory) > 0 {
		if len(s.ServerCert.PairName) == 0 {
			return fmt.Errorf("PairName is required if CertDirectory is set")
		}
		keyCert.CertFile = path.Join(s.ServerCert.CertDirectory, s.ServerCert.PairName+".crt")
		keyCert.KeyFile = path.Join(s.ServerCert.CertDirectory, s.ServerCert.PairName+".key")
		if canRead, err := certutil.CanReadCertAndKey(keyCert.CertFile, keyCert.KeyFile); err != nil {
			return err
		} else {
			canReadCertAndKey = canRead
		}
	}

	// 产生一个自签名的证书
	if !canReadCertAndKey {
		// 没设置就是 0.0.0.0
		if s.BindAddress.IsUnspecified() {
			alternateDNS = append(alternateDNS, "localhost")
		} else {
			alternateIPs = append(alternateIPs, s.BindAddress)
		}

		if cert, key, err := certutil.GenerateSelfSignedCertKeyWithFixtures(publicAddress, alternateIPs, alternateDNS, s.ServerCert.FixtureDirectory); err != nil {
			return fmt.Errorf("unable to generate self signed cert: %v", err)
		} else if len(keyCert.CertFile) > 0 && len(keyCert.KeyFile) > 0 {
			// 走到这里说明err为nil，cert和key都有值
			if err := certutil.WriteCert(keyCert.CertFile, cert); err != nil {
				return err
			}
			if err := keyutil.WriteKey(keyCert.KeyFile, key); err != nil {
				return err
			}
			klog.Infof("Generated self-signed cert (%s, %s)", keyCert.CertFile, keyCert.KeyFile)
		} else {
			panic("no impl")
		}
	}

	return nil
}

func CreateListener(network, addr string, config net.ListenConfig) (net.Listener, int, error) {
	if len(network) == 0 {
		network = "tcp"
	}

	ln, err := config.Listen(context.TODO(), network, addr)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to listen on %v: %v", addr, err)
	}

	// get port, 必须是tcp
	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		ln.Close()
		return nil, 0, fmt.Errorf("invalid listen address: %q", ln.Addr().String())
	}

	return ln, tcpAddr.Port, nil
}

// p375
type multipleControls []func(network, addr string, conn syscall.RawConn) error

func (mcs multipleControls) Control(network, addr string, conn syscall.RawConn) error {
	for _, c := range mcs {
		if err := c(network, addr, conn); err != nil {
			return err
		}
	}
	return nil
}
