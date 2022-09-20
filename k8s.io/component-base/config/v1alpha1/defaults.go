package v1alpha1

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

func RecommendedLoggingConfiguration(obj *LoggingConfiguration) {
	if obj.Format == "" {
		obj.Format = "text"
	}
	var empty resource.QuantityValue
	if obj.Options.JSON.InfoBufferSize == empty {
		obj.Options.JSON.InfoBufferSize = resource.QuantityValue{
			Quantity: *resource.NewQuantity(0, resource.DecimalSI),
		}
		_ = obj.Options.JSON.InfoBufferSize.String()
	}
	if obj.FlushFrequency == 0 {
		obj.FlushFrequency = 5 * time.Second
	}
}
