package validation

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const IsNegativeErrorMsg string = `must be greater than or equal to 0`

type ValidateNameFunc func(name string, prefix bool) []string

func NameIsDNSSubdomain(name string, prefix bool) []string {
	if prefix {
		name = maskTrailingDash(name)
	}
	return validation.IsDNS1123Subdomain(name)
}

func NameIsDNSLabel(name string, prefix bool) []string {
	if prefix {
		name = maskTrailingDash(name)
	}
	return validation.IsDNS1123Label(name)
}

func NameIsDNS1035Label(name string, prefix bool) []string {
	if prefix {
		name = maskTrailingDash(name)
	}
	return validation.IsDNS1035Label(name)
}

var ValidateNamespaceName = NameIsDNSLabel

var ValidateServiceAccountName = NameIsDNSSubdomain

func maskTrailingDash(name string) string {
	if len(name) > 1 && strings.HasSuffix(name, "-") {
		return name[:len(name)-2] + "a"
	}
	return name
}

func ValidateNonnegativeField(value int64, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if value < 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, value, IsNegativeErrorMsg))
	}
	return allErrs
}
