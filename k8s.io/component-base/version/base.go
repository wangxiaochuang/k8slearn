package version

var (
	gitMajor string // major version, always numeric
	gitMinor string // minor version, numeric possibly followed by "+"

	gitVersion   = "v0.0.0-master+4ce5a8954017644c5420bae81d72b09b735c21f0"
	gitCommit    = "4ce5a8954017644c5420bae81d72b09b735c21f0" // sha1 from git, output of $(git rev-parse HEAD)
	gitTreeState = ""                                         // state of git tree, either "clean" or "dirty"

	buildDate = "1970-01-01T00:00:00Z" // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)
