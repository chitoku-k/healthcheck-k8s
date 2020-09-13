package k8s

import (
	"context"
	"fmt"

	"github.com/chitoku-k/healthcheck-k8s/service"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type healthCheckService struct {
	Clientset kubernetes.Interface
}

func NewHealthCheckService(clientset kubernetes.Interface) service.HealthCheck {
	return &healthCheckService{
		Clientset: clientset,
	}
}

func (s *healthCheckService) Do(ctx context.Context, nodeName string) (bool, error) {
	node, err := s.Clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false, service.ErrNotFound
	}
	if err != nil {
		return false, fmt.Errorf(`failed to get node "%s": %w`, nodeName, err)
	}

	return !node.Spec.Unschedulable, nil
}
