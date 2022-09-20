package net

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type IPFamily string

const (
	IPv4 IPFamily = "4"
	IPv6          = "6"
)

type Protocol string

const (
	TCP Protocol = "TCP"
	UDP Protocol = "UDP"
)

type LocalPort struct {
	Description string
	IP          string
	IPFamily    IPFamily
	Port        int
	Protocol    Protocol
}

func NewLocalPort(desc, ip string, ipFamily IPFamily, port int, protocol Protocol) (*LocalPort, error) {
	if protocol != TCP && protocol != UDP {
		return nil, fmt.Errorf("Unsupported protocol %s", protocol)
	}
	if ipFamily != "" && ipFamily != "4" && ipFamily != "6" {
		return nil, fmt.Errorf("Invalid IP family %s", ipFamily)
	}
	if ip != "" {
		parsedIP := ParseIPSloppy(ip)
		if parsedIP == nil {
			return nil, fmt.Errorf("invalid ip address %s", ip)
		}
		asIPv4 := parsedIP.To4()
		if asIPv4 == nil && ipFamily == IPv4 || asIPv4 != nil && ipFamily == IPv6 {
			return nil, fmt.Errorf("ip address and family mismatch %s, %s", ip, ipFamily)
		}
	}
	return &LocalPort{Description: desc, IP: ip, IPFamily: ipFamily, Port: port, Protocol: protocol}, nil
}

func (lp *LocalPort) String() string {
	ipPort := net.JoinHostPort(lp.IP, strconv.Itoa(lp.Port))
	return fmt.Sprintf("%q (%s/%s%s)", lp.Description, ipPort, strings.ToLower(string(lp.Protocol)), lp.IPFamily)
}

type Closeable interface {
	Close() error
}

type PortOpener interface {
	OpenLocalPort(lp *LocalPort) (Closeable, error)
}

type listenPortOpener struct{}

func (l *listenPortOpener) OpenLocalPort(lp *LocalPort) (Closeable, error) {
	return openLocalPort(lp)
}

func openLocalPort(lp *LocalPort) (Closeable, error) {
	var socket Closeable
	hostPort := net.JoinHostPort(lp.IP, strconv.Itoa(lp.Port))
	switch lp.Protocol {
	case TCP:
		network := "tcp" + string(lp.IPFamily)
		listener, err := net.Listen(network, hostPort)
		if err != nil {
			return nil, err
		}
		socket = listener
	case UDP:
		network := "udp" + string(lp.IPFamily)
		addr, err := net.ResolveUDPAddr(network, hostPort)
		if err != nil {
			return nil, err
		}
		conn, err := net.ListenUDP(network, addr)
		if err != nil {
			return nil, err
		}
		socket = conn
	default:
		return nil, fmt.Errorf("unknown protocol %q", lp.Protocol)
	}
	return socket, nil
}
