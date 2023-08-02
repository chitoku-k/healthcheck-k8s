package k8s

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/chitoku-k/healthcheck-k8s/service"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	listercorev1 "k8s.io/client-go/listers/core/v1"
)

type healthCheckService struct {
	nodeLister listercorev1.NodeLister
}

func NewHealthCheckService(nodeLister listercorev1.NodeLister) service.HealthCheck {
	return &healthCheckService{
		nodeLister: nodeLister,
	}
}

func (s *healthCheckService) Do(ctx context.Context, nodeName string) (bool, error) {
	node, err := s.nodeLister.Get(nodeName)
	if apierrors.IsNotFound(err) {
		return false, service.NewNotFoundError(err)
	}
	if isTimeout(err) {
		return false, service.NewTimeoutError(err)
	}
	if err != nil {
		return false, fmt.Errorf(`failed to get node "%s": %w`, nodeName, err)
	}

	return !node.Spec.Unschedulable, nil
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
