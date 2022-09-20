package clientgo

import (
	_ "k8s.io/component-base/metrics/prometheus/clientgo/leaderelection" // load leaderelection metrics
	_ "k8s.io/component-base/metrics/prometheus/restclient"              // load restclient metrics
	_ "k8s.io/component-base/metrics/prometheus/workqueue"               // load the workqueue metrics
)
