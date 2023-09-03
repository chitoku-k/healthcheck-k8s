package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/chitoku-k/healthcheck-k8s/application/server"
	"github.com/chitoku-k/healthcheck-k8s/infrastructure/config"
	"github.com/chitoku-k/healthcheck-k8s/infrastructure/k8s"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	signals = []os.Signal{os.Interrupt}
	name    = "healthcheck-k8s"
	version = "v0.0.0-dev"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), signals...)
	defer stop()

	env, err := config.Get()
	if err != nil {
		slog.Error("Failed to initialize config", slog.Any("err", err))
		os.Exit(1)
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
		slog.Error("Failed to initialize kubeconfig", slog.Any("err", err))
		os.Exit(1)
	}

	config.Timeout = env.Timeout
	config.UserAgent = name + "/" + version

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error("Failed to initialize clientset", slog.Any("err", err))
		os.Exit(1)
	}

	informerFactory := informers.NewSharedInformerFactory(clientset, 0)
	nodeLister := informerFactory.Core().V1().Nodes().Lister()

	informerFactory.Start(ctx.Done())

	err = waitForCacheSync(ctx, env.Timeout, informerFactory)
	if err != nil {
		slog.Error("Failed to initialize node cache", slog.Any("err", err))
		os.Exit(1)
	}

	healthCheck := k8s.NewHealthCheckService(nodeLister)
	engine := server.NewEngine(env.Port, env.HeaderName, env.TrustedProxies, healthCheck)
	err = engine.Start(ctx)
	if err != nil {
		slog.Error("Failed to start web server", slog.Any("err", err))
		os.Exit(1)
	}
}

func waitForCacheSync(ctx context.Context, timeout time.Duration, factory informers.SharedInformerFactory) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for typ, done := range factory.WaitForCacheSync(ctx.Done()) {
		if !done {
			select {
			case <-ctx.Done():
				return fmt.Errorf("failed to sync %v: %v", typ, ctx.Err())
			default:
				return fmt.Errorf("failed to sync %v", typ)
			}
		}
	}

	return nil
}
