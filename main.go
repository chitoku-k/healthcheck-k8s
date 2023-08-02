package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/chitoku-k/healthcheck-k8s/application/server"
	"github.com/chitoku-k/healthcheck-k8s/infrastructure/config"
	"github.com/chitoku-k/healthcheck-k8s/infrastructure/k8s"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var signals = []os.Signal{os.Interrupt}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), signals...)
	defer stop()

	env, err := config.Get()
	if err != nil {
		panic(fmt.Errorf("failed to initialize config: %w", err))
	}

	var config *rest.Config
	kubeconfigPath := clientcmd.NewDefaultPathOptions().GetDefaultFilename()

	_, err = os.Stat(kubeconfigPath)
	if !os.IsNotExist(err) {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		panic(fmt.Errorf("failed to initialize kubeconfig: %w", err))
	}

	config.Timeout = env.Timeout
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Errorf("failed to initialize clientset: %w", err))
	}

	healthCheck := k8s.NewHealthCheckService(clientset)
	engine := server.NewEngine(env.Port, env.HeaderName, env.TrustedProxies, healthCheck)
	err = engine.Start(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to start web server: %v", err))
	}
}
