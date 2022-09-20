package metrics

import (
	"fmt"
	"regexp"

	"github.com/blang/semver/v4"

	apimachineryversion "k8s.io/apimachinery/pkg/version"
)

const (
	versionRegexpString = `^v(\d+\.\d+\.\d+)`
)

var (
	versionRe = regexp.MustCompile(versionRegexpString)
)

func parseSemver(s string) *semver.Version {
	if s != "" {
		sv := semver.MustParse(s)
		return &sv
	}
	return nil
}
func parseVersion(ver apimachineryversion.Info) semver.Version {
	matches := versionRe.FindAllStringSubmatch(ver.String(), -1)

	if len(matches) != 1 {
		panic(fmt.Sprintf("version string \"%v\" doesn't match expected regular expression: \"%v\"", ver.String(), versionRe.String()))
	}
	return semver.MustParse(matches[0][1])
}
