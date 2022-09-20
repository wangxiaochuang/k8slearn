package net

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
	"strconv"
)

func ParseCIDRs(cidrsString []string) ([]*net.IPNet, error) {
	cidrs := make([]*net.IPNet, 0, len(cidrsString))
	for _, cidrString := range cidrsString {
		_, cidr, err := ParseCIDRSloppy(cidrString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse cidr value:%q with error:%v", cidrString, err)
		}
		cidrs = append(cidrs, cidr)
	}
	return cidrs, nil
}

func IsDualStackIPs(ips []net.IP) (bool, error) {
	v4Found := false
	v6Found := false
	for _, ip := range ips {
		if ip == nil {
			return false, fmt.Errorf("ip %v is invalid", ip)
		}

		if v4Found && v6Found {
			continue
		}

		if IsIPv6(ip) {
			v6Found = true
			continue
		}

		v4Found = true
	}

	return (v4Found && v6Found), nil
}

func IsDualStackIPStrings(ips []string) (bool, error) {
	parsedIPs := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		parsedIP := ParseIPSloppy(ip)
		parsedIPs = append(parsedIPs, parsedIP)
	}
	return IsDualStackIPs(parsedIPs)
}

func IsDualStackCIDRs(cidrs []*net.IPNet) (bool, error) {
	v4Found := false
	v6Found := false
	for _, cidr := range cidrs {
		if cidr == nil {
			return false, fmt.Errorf("cidr %v is invalid", cidr)
		}

		if v4Found && v6Found {
			continue
		}

		if IsIPv6(cidr.IP) {
			v6Found = true
			continue
		}
		v4Found = true
	}

	return v4Found && v6Found, nil
}

func IsDualStackCIDRStrings(cidrs []string) (bool, error) {
	parsedCIDRs, err := ParseCIDRs(cidrs)
	if err != nil {
		return false, err
	}
	return IsDualStackCIDRs(parsedCIDRs)
}

func IsIPv6(netIP net.IP) bool {
	return netIP != nil && netIP.To4() == nil
}

func IsIPv6String(ip string) bool {
	netIP := ParseIPSloppy(ip)
	return IsIPv6(netIP)
}

func IsIPv6CIDRString(cidr string) bool {
	ip, _, _ := ParseCIDRSloppy(cidr)
	return IsIPv6(ip)
}

func IsIPv6CIDR(cidr *net.IPNet) bool {
	ip := cidr.IP
	return IsIPv6(ip)
}

func IsIPv4(netIP net.IP) bool {
	return netIP != nil && netIP.To4() != nil
}

func IsIPv4String(ip string) bool {
	netIP := ParseIPSloppy(ip)
	return IsIPv4(netIP)
}

func IsIPv4CIDR(cidr *net.IPNet) bool {
	ip := cidr.IP
	return IsIPv4(ip)
}

func IsIPv4CIDRString(cidr string) bool {
	ip, _, _ := ParseCIDRSloppy(cidr)
	return IsIPv4(ip)
}

func ParsePort(port string, allowZero bool) (int, error) {
	portInt, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, err
	}
	if portInt == 0 && !allowZero {
		return 0, errors.New("0 is not a valid port number")
	}
	return int(portInt), nil
}

func BigForIP(ip net.IP) *big.Int {
	return big.NewInt(0).SetBytes(ip.To16())
}

func AddIPOffset(base *big.Int, offset int) net.IP {
	r := big.NewInt(0).Add(base, big.NewInt(int64(offset))).Bytes()
	r = append(make([]byte, 16), r...)
	return net.IP(r[len(r)-16:])
}

func RangeSize(subnet *net.IPNet) int64 {
	ones, bits := subnet.Mask.Size()
	if bits == 32 && (bits-ones) >= 31 || bits == 128 && (bits-ones) >= 127 {
		return 0
	}
	if bits-ones >= 63 {
		return math.MaxInt64
	}

	return int64(1) << uint(bits-ones)
}

func GetIndexedIP(subnet *net.IPNet, index int) (net.IP, error) {
	ip := AddIPOffset(BigForIP(subnet.IP), index)
	if !subnet.Contains(ip) {
		return nil, fmt.Errorf("can't generate IP with index %d from subnet. subnet too small. subnet: %q", index, subnet)
	}
	return ip, nil
}
