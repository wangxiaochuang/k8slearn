package naming

import (
	"fmt"
	"regexp"
	goruntime "runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

func GetNameFromCallsite(ignoredPackages ...string) string {
	name := "????"
	const maxStack = 10
	for i := 1; i < maxStack; i++ {
		_, file, line, ok := goruntime.Caller(i)
		if !ok {
			file, line, ok = extractStackCreator()
			if !ok {
				break
			}
			i += maxStack
		}
		if hasPackage(file, append(ignoredPackages, "/runtime/asm_")) {
			continue
		}

		file = trimPackagePrefix(file)
		name = fmt.Sprintf("%s:%d", file, line)
		break
	}
	return name
}

func hasPackage(file string, ignoredPackages []string) bool {
	for _, ignoredPackage := range ignoredPackages {
		if strings.Contains(file, ignoredPackage) {
			return true
		}
	}
	return false
}

func trimPackagePrefix(file string) string {
	if l := strings.LastIndex(file, "/vendor/"); l >= 0 {
		return file[l+len("/vendor/"):]
	}
	if l := strings.LastIndex(file, "/src/"); l >= 0 {
		return file[l+5:]
	}
	if l := strings.LastIndex(file, "/pkg/"); l >= 0 {
		return file[l+1:]
	}
	return file
}

var stackCreator = regexp.MustCompile(`(?m)^created by (.*)\n\s+(.*):(\d+) \+0x[[:xdigit:]]+$`)

func extractStackCreator() (string, int, bool) {
	stack := debug.Stack()
	matches := stackCreator.FindStringSubmatch(string(stack))
	if len(matches) != 4 {
		return "", 0, false
	}
	line, err := strconv.Atoi(matches[3])
	if err != nil {
		return "", 0, false
	}
	return matches[2], line, true
}
