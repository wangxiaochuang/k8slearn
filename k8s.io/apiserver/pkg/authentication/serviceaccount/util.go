package serviceaccount

import (
	"fmt"
	"strings"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apiserver/pkg/authentication/user"
)

const (
	ServiceAccountUsernamePrefix    = "system:serviceaccount:"
	ServiceAccountUsernameSeparator = ":"
	ServiceAccountGroupPrefix       = "system:serviceaccounts:"
	AllServiceAccountsGroup         = "system:serviceaccounts"

	PodNameKey = "authentication.kubernetes.io/pod-name"
	PodUIDKey  = "authentication.kubernetes.io/pod-uid"
)

func MakeUsername(namespace, name string) string {
	return ServiceAccountUsernamePrefix + namespace + ServiceAccountUsernameSeparator + name
}

func MatchesUsername(namespace, name string, username string) bool {
	if !strings.HasPrefix(username, ServiceAccountUsernamePrefix) {
		return false
	}
	username = username[len(ServiceAccountUsernamePrefix):]

	if !strings.HasPrefix(username, namespace) {
		return false
	}
	username = username[len(namespace):]

	if !strings.HasPrefix(username, ServiceAccountUsernameSeparator) {
		return false
	}
	username = username[len(ServiceAccountUsernameSeparator):]

	return username == name
}

var invalidUsernameErr = fmt.Errorf("Username must be in the form %s", MakeUsername("namespace", "name"))

func SplitUsername(username string) (string, string, error) {
	if !strings.HasPrefix(username, ServiceAccountUsernamePrefix) {
		return "", "", invalidUsernameErr
	}
	trimmed := strings.TrimPrefix(username, ServiceAccountUsernamePrefix)
	parts := strings.Split(trimmed, ServiceAccountUsernameSeparator)
	if len(parts) != 2 {
		return "", "", invalidUsernameErr
	}
	namespace, name := parts[0], parts[1]
	if len(apimachineryvalidation.ValidateNamespaceName(namespace, false)) != 0 {
		return "", "", invalidUsernameErr
	}
	if len(apimachineryvalidation.ValidateServiceAccountName(name, false)) != 0 {
		return "", "", invalidUsernameErr
	}
	return namespace, name, nil
}

func MakeGroupNames(namespace string) []string {
	return []string{
		AllServiceAccountsGroup,
		MakeNamespaceGroupName(namespace),
	}
}

func MakeNamespaceGroupName(namespace string) string {
	return ServiceAccountGroupPrefix + namespace
}

func UserInfo(namespace, name, uid string) user.Info {
	return (&ServiceAccountInfo{
		Name:      name,
		Namespace: namespace,
		UID:       uid,
	}).UserInfo()
}

type ServiceAccountInfo struct {
	Name, Namespace, UID string
	PodName, PodUID      string
}

func (sa *ServiceAccountInfo) UserInfo() user.Info {
	info := &user.DefaultInfo{
		Name:   MakeUsername(sa.Namespace, sa.Name),
		UID:    sa.UID,
		Groups: MakeGroupNames(sa.Namespace),
	}
	if sa.PodName != "" && sa.PodUID != "" {
		info.Extra = map[string][]string{
			PodNameKey: {sa.PodName},
			PodUIDKey:  {sa.PodUID},
		}
	}
	return info
}
