package main

import (
	"fmt"

	"github.com/chitoku-k/healthcheck-k8s/application/server"
	"github.com/chitoku-k/healthcheck-k8s/infrastructure/config"
	"github.com/chitoku-k/healthcheck-k8s/infrastructure/k8s"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	env, err := config.Get()
	if err != nil {
		panic(fmt.Errorf("failed to initialize config: %w", err))
	}

	var config *rest.Config
	if env.KubeConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", env.KubeConfig)
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
	engine.Start()
}
