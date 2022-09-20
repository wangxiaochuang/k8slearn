package net

import (
	"fmt"
	"net"
	"strings"
)

type IPNetSet map[string]*net.IPNet

func ParseIPNets(specs ...string) (IPNetSet, error) {
	ipnetset := make(IPNetSet)
	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		_, ipnet, err := ParseCIDRSloppy(spec)
		if err != nil {
			return nil, err
		}
		k := ipnet.String()
		ipnetset[k] = ipnet
	}
	return ipnetset, nil
}

func (s IPNetSet) Insert(items ...*net.IPNet) {
	for _, item := range items {
		s[item.String()] = item
	}
}

func (s IPNetSet) Delete(items ...*net.IPNet) {
	for _, item := range items {
		delete(s, item.String())
	}
}

func (s IPNetSet) Has(item *net.IPNet) bool {
	_, contained := s[item.String()]
	return contained
}

func (s IPNetSet) HasAll(items ...*net.IPNet) bool {
	for _, item := range items {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

func (s IPNetSet) Difference(s2 IPNetSet) IPNetSet {
	result := make(IPNetSet)
	for k, i := range s {
		_, found := s2[k]
		if found {
			continue
		}
		result[k] = i
	}
	return result
}

func (s IPNetSet) StringSlice() []string {
	a := make([]string, 0, len(s))
	for k := range s {
		a = append(a, k)
	}
	return a
}

func (s IPNetSet) IsSuperset(s2 IPNetSet) bool {
	for k := range s2 {
		_, found := s[k]
		if !found {
			return false
		}
	}
	return true
}

func (s IPNetSet) Equal(s2 IPNetSet) bool {
	return len(s) == len(s2) && s.IsSuperset(s2)
}

func (s IPNetSet) Len() int {
	return len(s)
}

type IPSet map[string]net.IP

func ParseIPSet(items ...string) (IPSet, error) {
	ipset := make(IPSet)
	for _, item := range items {
		ip := ParseIPSloppy(strings.TrimSpace(item))
		if ip == nil {
			return nil, fmt.Errorf("error parsing IP %q", item)
		}

		ipset[ip.String()] = ip
	}

	return ipset, nil
}

func (s IPSet) Insert(items ...net.IP) {
	for _, item := range items {
		s[item.String()] = item
	}
}

func (s IPSet) Delete(items ...net.IP) {
	for _, item := range items {
		delete(s, item.String())
	}
}

func (s IPSet) Has(item net.IP) bool {
	_, contained := s[item.String()]
	return contained
}

func (s IPSet) HasAll(items ...net.IP) bool {
	for _, item := range items {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

func (s IPSet) Difference(s2 IPSet) IPSet {
	result := make(IPSet)
	for k, i := range s {
		_, found := s2[k]
		if found {
			continue
		}
		result[k] = i
	}
	return result
}

func (s IPSet) StringSlice() []string {
	a := make([]string, 0, len(s))
	for k := range s {
		a = append(a, k)
	}
	return a
}

func (s IPSet) IsSuperset(s2 IPSet) bool {
	for k := range s2 {
		_, found := s[k]
		if !found {
			return false
		}
	}
	return true
}

func (s IPSet) Equal(s2 IPSet) bool {
	return len(s) == len(s2) && s.IsSuperset(s2)
}

func (s IPSet) Len() int {
	return len(s)
}
