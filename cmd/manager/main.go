package main

import (
	"flag"
	"github.com/j-vizcaino/k8s-controller-datadog-monitor/pkg/datadog-client"
	"os"
	"runtime"
	log2 "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/j-vizcaino/k8s-controller-datadog-monitor/pkg/apis"
	"github.com/j-vizcaino/k8s-controller-datadog-monitor/pkg/controller"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func addDatadogClient() error {

	clt, err := datadog_client.NewClient("api.datadoghq.com",
		os.Getenv("DD_API_KEY"),
		os.Getenv("DD_APP_KEY"))
	if err != nil {
		return err
	}

	datadog_client.Registry[clt.Host()] = clt
	return nil
}


func main() {
	logger := log2.ZapLogger(true)
	log2.SetLogger(logger)

	logger.Info("Monitor controller", "go_version", runtime.Version(), "operator_sdk", sdkVersion.Version)
	flag.Parse()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logger.Error(err, "failed to get watch namespace")
		os.Exit(1)
	}

	// TODO: Expose metrics port after SDK uses controller-runtime's dynamic client
	// sdk.ExposeMetricsPort()

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error(err, "Failed to get config to talk to Kubernetes API server")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{Namespace: namespace})
	if err != nil {
		logger.Error(err, "Manager creation failed")
		os.Exit(1)
	}

	logger.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		logger.Error(err, "API schemes registration failed")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		logger.Error(err, "Controller setup failed")
		os.Exit(1)
	}

	if err := addDatadogClient(); err != nil {
		logger.Error(err, "Failed to intialize Datadog API client")
		os.Exit(1)
	}

	logger.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logger.Error(err, "Manager start failed")
		os.Exit(1)
	}
}
