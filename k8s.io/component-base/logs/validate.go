package logs

import (
	"fmt"
	"math"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/config"
	"k8s.io/component-base/logs/registry"
)

func ValidateLoggingConfiguration(c *config.LoggingConfiguration, fldPath *field.Path) field.ErrorList {
	errs := field.ErrorList{}
	if c.Format != DefaultLogFormat {
		allFlags := UnsupportedLoggingFlags(cliflag.WordSepNormalizeFunc)
		// 不支持的flag都必须是默认值
		for _, f := range allFlags {
			if f.DefValue != f.Value.String() {
				errs = append(errs, field.Invalid(fldPath.Child("format"), c.Format, fmt.Sprintf("Non-default format doesn't honor flag: %s", f.Name)))
			}
		}
	}
	_, err := registry.LogRegistry.Get(c.Format)
	if err != nil {
		errs = append(errs, field.Invalid(fldPath.Child("format"), c.Format, "Unsupported log format"))
	}

	if c.Verbosity > math.MaxInt32 {
		errs = append(errs, field.Invalid(fldPath.Child("verbosity"), c.Verbosity, fmt.Sprintf("Must be <= %d", math.MaxInt32)))
	}
	vmoduleFldPath := fldPath.Child("vmodule")
	if len(c.VModule) > 0 && c.Format != "" && c.Format != "text" {
		errs = append(errs, field.Forbidden(vmoduleFldPath, "Only supported for text log format"))
	}

	for i, item := range c.VModule {
		if item.FilePattern == "" {
			errs = append(errs, field.Required(vmoduleFldPath.Index(i), "File pattern must not be empty"))
		}
		if strings.ContainsAny(item.FilePattern, "=,") {
			errs = append(errs, field.Invalid(vmoduleFldPath.Index(i), item.FilePattern, "File pattern must not contain equal sign or comma"))
		}
		if item.Verbosity > math.MaxInt32 {
			errs = append(errs, field.Invalid(vmoduleFldPath.Index(i), item.Verbosity, fmt.Sprintf("Must be <= %d", math.MaxInt32)))
		}
	}

	return errs
}
