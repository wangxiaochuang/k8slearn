package net

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/http2"
	"k8s.io/klog/v2"
	netutils "k8s.io/utils/net"
)

// 应对 ["", "/foo/"]这种，join会去掉最后一个/，但是要保留
func JoinPreservingTrailingSlash(elem ...string) string {
	result := path.Join(elem...)

	for i := len(elem) - 1; i >= 0; i-- {
		if len(elem[i]) > 0 {
			if strings.HasSuffix(elem[i], "/") && !strings.HasSuffix(result, "/") {
				result += "/"
			}
			break
		}
	}

	return result
}

func IsTimeout(err error) bool {
	var neterr net.Error
	if errors.As(err, &neterr) {
		return neterr != nil && neterr.Timeout()
	}
	return false
}

func IsProbableEOF(err error) bool {
	if err == nil {
		return false
	}
	var uerr *url.Error
	if errors.As(err, &uerr) {
		err = uerr.Err
	}
	msg := err.Error()
	switch {
	case err == io.EOF:
		return true
	case err == io.ErrUnexpectedEOF:
		return true
	case msg == "http: can't write HTTP request on broken connection":
		return true
	case strings.Contains(msg, "http2: server sent GOAWAY and closed the connection"):
		return true
	case strings.Contains(msg, "connection reset by peer"):
		return true
	case strings.Contains(strings.ToLower(msg), "use of closed network connection"):
		return true
	}
	return false
}

var defaultTransport = http.DefaultTransport.(*http.Transport)

func SetOldTransportDefaults(t *http.Transport) *http.Transport {
	if t.Proxy == nil || isDefault(t.Proxy) {
		t.Proxy = NewProxierWithNoProxyCIDR(http.ProxyFromEnvironment)
	}
	if t.DialContext == nil && t.Dial == nil {
		t.DialContext = defaultTransport.DialContext
	}
	if t.TLSHandshakeTimeout == 0 {
		t.TLSHandshakeTimeout = defaultTransport.TLSHandshakeTimeout
	}
	if t.IdleConnTimeout == 0 {
		t.IdleConnTimeout = defaultTransport.IdleConnTimeout
	}
	return t
}

func SetTransportDefaults(t *http.Transport) *http.Transport {
	t = SetOldTransportDefaults(t)
	if s := os.Getenv("DISABLE_HTTP2"); len(s) > 0 {
		klog.Info("HTTP2 has been explicitly disabled")
	} else if allowsHTTP2(t) {
		if err := configureHTTP2Transport(t); err != nil {
			klog.Warningf("Transport failed http2 configuration: %v", err)
		}
	}
	return t
}

func readIdleTimeoutSeconds() int {
	ret := 30
	// User can set the readIdleTimeout to 0 to disable the HTTP/2
	// connection health check.
	if s := os.Getenv("HTTP2_READ_IDLE_TIMEOUT_SECONDS"); len(s) > 0 {
		i, err := strconv.Atoi(s)
		if err != nil {
			klog.Warningf("Illegal HTTP2_READ_IDLE_TIMEOUT_SECONDS(%q): %v."+
				" Default value %d is used", s, err, ret)
			return ret
		}
		ret = i
	}
	return ret
}

func pingTimeoutSeconds() int {
	ret := 15
	if s := os.Getenv("HTTP2_PING_TIMEOUT_SECONDS"); len(s) > 0 {
		i, err := strconv.Atoi(s)
		if err != nil {
			klog.Warningf("Illegal HTTP2_PING_TIMEOUT_SECONDS(%q): %v."+
				" Default value %d is used", s, err, ret)
			return ret
		}
		ret = i
	}
	return ret
}

func configureHTTP2Transport(t *http.Transport) error {
	t2, err := http2.ConfigureTransports(t)
	if err != nil {
		return err
	}
	t2.ReadIdleTimeout = time.Duration(readIdleTimeoutSeconds()) * time.Second
	t2.PingTimeout = time.Duration(pingTimeoutSeconds()) * time.Second
	return nil
}

func allowsHTTP2(t *http.Transport) bool {
	if t.TLSClientConfig == nil || len(t.TLSClientConfig.NextProtos) == 0 {
		return true
	}
	for _, p := range t.TLSClientConfig.NextProtos {
		if p == http2.NextProtoTLS {
			return true
		}
	}
	return false
}

type RoundTripperWrapper interface {
	http.RoundTripper
	WrappedRoundTripper() http.RoundTripper
}

type DialFunc func(ctx context.Context, net, addr string) (net.Conn, error)

func DialerFor(transport http.RoundTripper) (DialFunc, error) {
	if transport == nil {
		return nil, nil
	}

	switch transport := transport.(type) {
	case *http.Transport:
		if transport.DialContext != nil {
			return transport.DialContext, nil
		}
		if transport.Dial != nil {
			return func(ctx context.Context, net, addr string) (net.Conn, error) {
				return transport.Dial(net, addr)
			}, nil
		}
		return nil, nil
	case RoundTripperWrapper:
		return DialerFor(transport.WrappedRoundTripper())
	default:
		return nil, fmt.Errorf("unknown transport type: %T", transport)
	}
}

func CloseIdleConnectionsFor(transport http.RoundTripper) {
	if transport == nil {
		return
	}
	type closeIdler interface {
		CloseIdleConnections()
	}

	switch transport := transport.(type) {
	case closeIdler:
		transport.CloseIdleConnections()
	case RoundTripperWrapper:
		CloseIdleConnectionsFor(transport.WrappedRoundTripper())
	default:
		klog.Warningf("unknown transport type: %T", transport)
	}
}

type TLSClientConfigHolder interface {
	TLSClientConfig() *tls.Config
}

func TLSClientConfig(transport http.RoundTripper) (*tls.Config, error) {
	if transport == nil {
		return nil, nil
	}

	switch transport := transport.(type) {
	case *http.Transport:
		return transport.TLSClientConfig, nil
	case TLSClientConfigHolder:
		return transport.TLSClientConfig(), nil
	case RoundTripperWrapper:
		return TLSClientConfig(transport.WrappedRoundTripper())
	default:
		return nil, fmt.Errorf("unknown transport type: %T", transport)
	}
}

func FormatURL(scheme string, host string, port int, path string) *url.URL {
	return &url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
		Path:   path,
	}
}

func GetHTTPClient(req *http.Request) string {
	if ua := req.UserAgent(); len(ua) != 0 {
		return ua
	}
	return "unknown"
}

func SourceIPs(req *http.Request) []net.IP {
	var srcIPs []net.IP

	hdr := req.Header
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrForwardedFor != "" {
		parts := strings.Split(hdrForwardedFor, ",")
		for _, part := range parts {
			ip := netutils.ParseIPSloppy(strings.TrimSpace(part))
			if ip != nil {
				srcIPs = append(srcIPs, ip)
			}
		}
	}

	hdrRealIp := hdr.Get("X-Real-Ip")
	if hdrRealIp != "" {
		ip := netutils.ParseIPSloppy(hdrRealIp)
		// Only append the X-Real-Ip if it's not already contained in the X-Forwarded-For chain.
		if ip != nil && !containsIP(srcIPs, ip) {
			srcIPs = append(srcIPs, ip)
		}
	}

	var remoteIP net.IP
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		remoteIP = netutils.ParseIPSloppy(host)
	}
	if remoteIP == nil {
		remoteIP = netutils.ParseIPSloppy(req.RemoteAddr)
	}
	if remoteIP != nil && (len(srcIPs) == 0 || !remoteIP.Equal(srcIPs[len(srcIPs)-1])) {
		srcIPs = append(srcIPs, remoteIP)
	}

	return srcIPs
}

func containsIP(ips []net.IP, ip net.IP) bool {
	for _, v := range ips {
		if v.Equal(ip) {
			return true
		}
	}
	return false
}

func GetClientIP(req *http.Request) net.IP {
	ips := SourceIPs(req)
	if len(ips) == 0 {
		return nil
	}
	return ips[0]
}

func AppendForwardedForHeader(req *http.Request) {
	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if prior, ok := req.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		req.Header.Set("X-Forwarded-For", clientIP)
	}
}

var defaultProxyFuncPointer = fmt.Sprintf("%p", http.ProxyFromEnvironment)

func isDefault(transportProxier func(*http.Request) (*url.URL, error)) bool {
	transportProxierPointer := fmt.Sprintf("%p", transportProxier)
	return transportProxierPointer == defaultProxyFuncPointer
}

func NewProxierWithNoProxyCIDR(delegate func(req *http.Request) (*url.URL, error)) func(req *http.Request) (*url.URL, error) {
	noProxyEnv := os.Getenv("NO_PROXY")
	if noProxyEnv == "" {
		noProxyEnv = os.Getenv("no_proxy")
	}
	noProxyRules := strings.Split(noProxyEnv, ",")

	cidrs := []*net.IPNet{}
	for _, noProxyRule := range noProxyRules {
		_, cidr, _ := netutils.ParseCIDRSloppy(noProxyRule)
		if cidr != nil {
			cidrs = append(cidrs, cidr)
		}
	}

	if len(cidrs) == 0 {
		return delegate
	}

	return func(req *http.Request) (*url.URL, error) {
		ip := netutils.ParseIPSloppy(req.URL.Hostname())
		if ip == nil {
			return delegate(req)
		}
		for _, cidr := range cidrs {
			if cidr.Contains(ip) {
				return nil, nil
			}
		}

		return delegate(req)
	}
}

type DialerFunc func(req *http.Request) (net.Conn, error)

func (fn DialerFunc) Dial(req *http.Request) (net.Conn, error) {
	return fn(req)
}

type Dialer interface {
	Dial(req *http.Request) (net.Conn, error)
}

func CloneRequest(req *http.Request) *http.Request {
	r := new(http.Request)

	*r = *req

	r.Header = CloneHeader(req.Header)

	return r
}

func CloneHeader(in http.Header) http.Header {
	out := make(http.Header, len(in))
	for key, values := range in {
		newValues := make([]string, len(values))
		copy(newValues, values)
		out[key] = newValues
	}
	return out
}

type WarningHeader struct {
	Code  int
	Agent string
	Text  string
}

func ParseWarningHeaders(headers []string) ([]WarningHeader, []error) {
	var (
		results []WarningHeader
		errs    []error
	)
	for _, header := range headers {
		for len(header) > 0 {
			result, remainder, err := ParseWarningHeader(header)
			if err != nil {
				errs = append(errs, err)
				break
			}
			results = append(results, result)
			header = remainder
		}
	}
	return results, errs
}

var (
	codeMatcher = regexp.MustCompile(`^[0-9]{3}$`)
	wordDecoder = &mime.WordDecoder{}
)

func ParseWarningHeader(header string) (result WarningHeader, remainder string, err error) {
	header = strings.TrimSpace(header)

	parts := strings.SplitN(header, " ", 3)
	if len(parts) != 3 {
		return WarningHeader{}, "", errors.New("invalid warning header: fewer than 3 segments")
	}
	code, agent, textDateRemainder := parts[0], parts[1], parts[2]

	if !codeMatcher.Match([]byte(code)) {
		return WarningHeader{}, "", errors.New("invalid warning header: code segment is not 3 digits between 100-299")
	}
	codeInt, _ := strconv.ParseInt(code, 10, 64)

	if len(agent) == 0 {
		return WarningHeader{}, "", errors.New("invalid warning header: empty agent segment")
	}
	if !utf8.ValidString(agent) || hasAnyRunes(agent, unicode.IsControl) {
		return WarningHeader{}, "", errors.New("invalid warning header: invalid agent")
	}

	if len(textDateRemainder) == 0 {
		return WarningHeader{}, "", errors.New("invalid warning header: empty text segment")
	}

	text, dateAndRemainder, err := parseQuotedString(textDateRemainder)
	if err != nil {
		return WarningHeader{}, "", fmt.Errorf("invalid warning header: %v", err)
	}

	if decodedText, err := wordDecoder.DecodeHeader(text); err == nil {
		text = decodedText
	}
	if !utf8.ValidString(text) || hasAnyRunes(text, unicode.IsControl) {
		return WarningHeader{}, "", errors.New("invalid warning header: invalid text")
	}

	result = WarningHeader{Code: int(codeInt), Agent: agent, Text: text}

	if len(dateAndRemainder) > 0 {
		if dateAndRemainder[0] == '"' {
			// consume date
			foundEndQuote := false
			for i := 1; i < len(dateAndRemainder); i++ {
				if dateAndRemainder[i] == '"' {
					foundEndQuote = true
					remainder = strings.TrimSpace(dateAndRemainder[i+1:])
					break
				}
			}
			if !foundEndQuote {
				return WarningHeader{}, "", errors.New("invalid warning header: unterminated date segment")
			}
		} else {
			remainder = dateAndRemainder
		}
	}
	if len(remainder) > 0 {
		if remainder[0] == ',' {
			// consume comma if present
			remainder = strings.TrimSpace(remainder[1:])
		} else {
			return WarningHeader{}, "", errors.New("invalid warning header: unexpected token after warn-date")
		}
	}

	return result, remainder, nil
}

func parseQuotedString(quotedString string) (string, string, error) {
	if len(quotedString) == 0 {
		return "", "", errors.New("invalid quoted string: 0-length")
	}

	if quotedString[0] != '"' {
		return "", "", errors.New("invalid quoted string: missing initial quote")
	}

	quotedString = quotedString[1:]
	var remainder string
	escaping := false
	closedQuote := false
	result := &strings.Builder{}
loop:
	for i := 0; i < len(quotedString); i++ {
		b := quotedString[i]
		switch b {
		case '"':
			if escaping {
				result.WriteByte(b)
				escaping = false
			} else {
				closedQuote = true
				remainder = strings.TrimSpace(quotedString[i+1:])
				break loop
			}
		case '\\':
			if escaping {
				result.WriteByte(b)
				escaping = false
			} else {
				escaping = true
			}
		default:
			result.WriteByte(b)
			escaping = false
		}
	}
	if !closedQuote {
		return "", "", errors.New("invalid quoted string: missing closing quote")
	}
	return result.String(), remainder, nil
}

func NewWarningHeader(code int, agent, text string) (string, error) {
	if code < 0 || code > 999 {
		return "", errors.New("code must be between 0 and 999")
	}
	if len(agent) == 0 {
		agent = "-"
	} else if !utf8.ValidString(agent) || strings.ContainsAny(agent, `\"`) || hasAnyRunes(agent, unicode.IsSpace, unicode.IsControl) {
		return "", errors.New("agent must be valid UTF-8 and must not contain spaces, quotes, backslashes, or control characters")
	}
	if !utf8.ValidString(text) || hasAnyRunes(text, unicode.IsControl) {
		return "", errors.New("text must be valid UTF-8 and must not contain control characters")
	}
	return fmt.Sprintf("%03d %s %s", code, agent, makeQuotedString(text)), nil
}

func hasAnyRunes(s string, runeCheckers ...func(rune) bool) bool {
	for _, r := range s {
		for _, checker := range runeCheckers {
			if checker(r) {
				return true
			}
		}
	}
	return false
}

func makeQuotedString(s string) string {
	result := &bytes.Buffer{}
	// opening quote
	result.WriteRune('"')
	for _, c := range s {
		switch c {
		case '"', '\\':
			// escape " and \
			result.WriteRune('\\')
			result.WriteRune(c)
		default:
			// write everything else as-is
			result.WriteRune(c)
		}
	}
	// closing quote
	result.WriteRune('"')
	return result.String()
}
