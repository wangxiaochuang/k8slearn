package net

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

var validSchemes = sets.NewString("http", "https", "")

func SplitSchemeNamePort(id string) (scheme, name, port string, valid bool) {
	parts := strings.Split(id, ":")
	switch len(parts) {
	case 1:
		name = parts[0]
	case 2:
		name = parts[0]
		port = parts[1]
	case 3:
		scheme = parts[0]
		name = parts[1]
		port = parts[2]
	default:
		return "", "", "", false
	}

	if len(name) > 0 && validSchemes.Has(scheme) {
		return scheme, name, port, true
	} else {
		return "", "", "", false
	}
}

func JoinSchemeNamePort(scheme, name, port string) string {
	if len(scheme) > 0 {
		return scheme + ":" + name + ":" + port
	}
	if len(port) > 0 {
		return name + ":" + port
	}
	return name
}
