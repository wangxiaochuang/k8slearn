package metrics

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	labelValueAllowLists = map[string]*MetricLabelAllowList{}
	allowListLock        sync.RWMutex
)

type KubeOpts struct {
	Namespace            string
	Subsystem            string
	Name                 string
	Help                 string
	ConstLabels          map[string]string
	DeprecatedVersion    string
	deprecateOnce        sync.Once
	annotateOnce         sync.Once
	StabilityLevel       StabilityLevel
	LabelValueAllowLists *MetricLabelAllowList
}

func BuildFQName(namespace, subsystem, name string) string {
	return prometheus.BuildFQName(namespace, subsystem, name)
}

type StabilityLevel string

const (
	ALPHA  StabilityLevel = "ALPHA"
	STABLE StabilityLevel = "STABLE"
)

func (sl *StabilityLevel) setDefaults() {
	switch *sl {
	case "":
		*sl = ALPHA
	default:
		// no-op, since we have a StabilityLevel already
	}
}

type CounterOpts KubeOpts

func (o *CounterOpts) markDeprecated() {
	o.deprecateOnce.Do(func() {
		o.Help = fmt.Sprintf("(Deprecated since %v) %v", o.DeprecatedVersion, o.Help)
	})
}

func (o *CounterOpts) annotateStabilityLevel() {
	o.annotateOnce.Do(func() {
		o.Help = fmt.Sprintf("[%v] %v", o.StabilityLevel, o.Help)
	})
}

func (o *CounterOpts) toPromCounterOpts() prometheus.CounterOpts {
	return prometheus.CounterOpts{
		Namespace:   o.Namespace,
		Subsystem:   o.Subsystem,
		Name:        o.Name,
		Help:        o.Help,
		ConstLabels: o.ConstLabels,
	}
}

type GaugeOpts KubeOpts

// Modify help description on the metric description.
func (o *GaugeOpts) markDeprecated() {
	o.deprecateOnce.Do(func() {
		o.Help = fmt.Sprintf("(Deprecated since %v) %v", o.DeprecatedVersion, o.Help)
	})
}

func (o *GaugeOpts) annotateStabilityLevel() {
	o.annotateOnce.Do(func() {
		o.Help = fmt.Sprintf("[%v] %v", o.StabilityLevel, o.Help)
	})
}

// convenience function to allow easy transformation to the prometheus
// counterpart. This will do more once we have a proper label abstraction
func (o *GaugeOpts) toPromGaugeOpts() prometheus.GaugeOpts {
	return prometheus.GaugeOpts{
		Namespace:   o.Namespace,
		Subsystem:   o.Subsystem,
		Name:        o.Name,
		Help:        o.Help,
		ConstLabels: o.ConstLabels,
	}
}

// p150
type HistogramOpts struct {
	Namespace            string
	Subsystem            string
	Name                 string
	Help                 string
	ConstLabels          map[string]string
	Buckets              []float64
	DeprecatedVersion    string
	deprecateOnce        sync.Once
	annotateOnce         sync.Once
	StabilityLevel       StabilityLevel
	LabelValueAllowLists *MetricLabelAllowList
}

// p165
func (o *HistogramOpts) markDeprecated() {
	o.deprecateOnce.Do(func() {
		o.Help = fmt.Sprintf("(Deprecated since %v) %v", o.DeprecatedVersion, o.Help)
	})
}

// p173
func (o *HistogramOpts) annotateStabilityLevel() {
	o.annotateOnce.Do(func() {
		o.Help = fmt.Sprintf("[%v] %v", o.StabilityLevel, o.Help)
	})
}

func (o *HistogramOpts) toPromHistogramOpts() prometheus.HistogramOpts {
	return prometheus.HistogramOpts{
		Namespace:   o.Namespace,
		Subsystem:   o.Subsystem,
		Name:        o.Name,
		Help:        o.Help,
		ConstLabels: o.ConstLabels,
		Buckets:     o.Buckets,
	}
}

type SummaryOpts struct {
	Namespace            string
	Subsystem            string
	Name                 string
	Help                 string
	ConstLabels          map[string]string
	Objectives           map[float64]float64
	MaxAge               time.Duration
	AgeBuckets           uint32
	BufCap               uint32
	DeprecatedVersion    string
	deprecateOnce        sync.Once
	annotateOnce         sync.Once
	StabilityLevel       StabilityLevel
	LabelValueAllowLists *MetricLabelAllowList
}

func (o *SummaryOpts) markDeprecated() {
	o.deprecateOnce.Do(func() {
		o.Help = fmt.Sprintf("(Deprecated since %v) %v", o.DeprecatedVersion, o.Help)
	})
}

func (o *SummaryOpts) annotateStabilityLevel() {
	o.annotateOnce.Do(func() {
		o.Help = fmt.Sprintf("[%v] %v", o.StabilityLevel, o.Help)
	})
}

var (
	defObjectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
)

func (o *SummaryOpts) toPromSummaryOpts() prometheus.SummaryOpts {
	// we need to retain existing quantile behavior for backwards compatibility,
	// so let's do what prometheus used to do prior to v1.
	objectives := o.Objectives
	if objectives == nil {
		objectives = defObjectives
	}
	return prometheus.SummaryOpts{
		Namespace:   o.Namespace,
		Subsystem:   o.Subsystem,
		Name:        o.Name,
		Help:        o.Help,
		ConstLabels: o.ConstLabels,
		Objectives:  objectives,
		MaxAge:      o.MaxAge,
		AgeBuckets:  o.AgeBuckets,
		BufCap:      o.BufCap,
	}
}

// p257
type MetricLabelAllowList struct {
	labelToAllowList map[string]sets.String
}

func (allowList *MetricLabelAllowList) ConstrainToAllowedList(labelNameList, labelValueList []string) {
	for index, value := range labelValueList {
		name := labelNameList[index]
		if allowValues, ok := allowList.labelToAllowList[name]; ok {
			if !allowValues.Has(value) {
				labelValueList[index] = "unexpected"
			}
		}
	}
}

func (allowList *MetricLabelAllowList) ConstrainLabelMap(labels map[string]string) {
	for name, value := range labels {
		if allowValues, ok := allowList.labelToAllowList[name]; ok {
			if !allowValues.Has(value) {
				labels[name] = "unexpected"
			}
		}
	}
}

func SetLabelAllowListFromCLI(allowListMapping map[string]string) {
	allowListLock.Lock()
	defer allowListLock.Unlock()
	for metricLabelName, labelValues := range allowListMapping {
		metricName := strings.Split(metricLabelName, ",")[0]
		labelName := strings.Split(metricLabelName, ",")[1]
		valueSet := sets.NewString(strings.Split(labelValues, ",")...)

		allowList, ok := labelValueAllowLists[metricName]
		if ok {
			allowList.labelToAllowList[labelName] = valueSet
		} else {
			labelToAllowList := make(map[string]sets.String)
			labelToAllowList[labelName] = valueSet
			labelValueAllowLists[metricName] = &MetricLabelAllowList{
				labelToAllowList,
			}
		}
	}
}
