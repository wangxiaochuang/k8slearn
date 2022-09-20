package strings

import "strings"

func EscapeQualifiedName(in string) string {
	return strings.Replace(in, "/", "~", -1)
}

func UnescapeQualifiedName(in string) string {
	return strings.Replace(in, "~", "/", -1)
}
