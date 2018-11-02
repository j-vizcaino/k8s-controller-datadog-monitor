package controller

import (
	"github.com/j-vizcaino/k8s-controller-datadog-monitor/pkg/controller/monitor"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, monitor.Add)
}
