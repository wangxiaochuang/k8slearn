package version

import (
	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/component-base/version"
)

var (
	buildInfo = metrics.NewGaugeVec(
		&metrics.GaugeOpts{
			Name:           "kubernetes_build_info",
			Help:           "A metric with a constant '1' value labeled by major, minor, git version, git commit, git tree state, build date, Go version, and compiler from which Kubernetes was built, and platform on which it is running.",
			StabilityLevel: metrics.ALPHA,
		},
		[]string{"major", "minor", "git_version", "git_commit", "git_tree_state", "build_date", "go_version", "compiler", "platform"},
	)
)

func init() {
	info := version.Get()
	legacyregistry.MustRegister(buildInfo)
	buildInfo.WithLabelValues(info.Major, info.Minor, info.GitVersion, info.GitCommit, info.GitTreeState, info.BuildDate, info.GoVersion, info.Compiler, info.Platform).Set(1)
}
