package options

import "k8s.io/client-go/util/keyutil"

// p231
func IsValidServiceAccountKeyFile(file string) bool {
	_, err := keyutil.PublicKeysFromFile(file)
	return err == nil
}
