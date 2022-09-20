package strings

import (
	"path"
	"strings"
)

func SplitQualifiedName(str string) (string, string) {
	parts := strings.Split(str, "/")
	if len(parts) < 2 {
		return "", str
	}
	return parts[0], parts[1]
}

func JoinQualifiedName(namespace, name string) string {
	return path.Join(namespace, name)
}

func ShortenString(str string, n int) string {
	if len(str) <= n {
		return str
	}
	return str[:n]
}
