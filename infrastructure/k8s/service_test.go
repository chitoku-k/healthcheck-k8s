package k8s_test

import (
	"github.com/chitoku-k/healthcheck-k8s/infrastructure/k8s"
	"github.com/chitoku-k/healthcheck-k8s/service"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applycorev1 "k8s.io/client-go/applyconfigurations/core/v1"
)

var _ = Describe("HealthCheckService", func() {
	var (
		node1ApplyConfiguration *applycorev1.NodeApplyConfiguration
		node2ApplyConfiguration *applycorev1.NodeApplyConfiguration
		healthCheckService      service.HealthCheck
	)

	BeforeEach(func() {
		node1ApplyConfiguration = applycorev1.Node("node-1").
			WithSpec(applycorev1.NodeSpec().
				WithUnschedulable(false))

		node2ApplyConfiguration = applycorev1.Node("node-2").
			WithSpec(applycorev1.NodeSpec().
				WithUnschedulable(true))

		healthCheckService = k8s.NewHealthCheckService(clientset)
	})

	Context("Do()", func() {
		Context("when node is not found", func() {
			It("returns NotFoundError", func(ctx SpecContext) {
				actual, err := healthCheckService.Do(ctx, "minikube")
				Expect(actual).To(BeFalse())
				Expect(err).To(MatchError(HavePrefix("not found:")))
			})
		})

		Context("when node is found", func() {
			Context("when node is schedulable", func() {
				BeforeEach(func(ctx SpecContext) {
					node1, err := clientset.CoreV1().Nodes().Apply(ctx, node1ApplyConfiguration, metav1.ApplyOptions{FieldManager: fieldManager})
					Expect(err).NotTo(HaveOccurred())

					node1.Status.Conditions = []corev1.NodeCondition{
						{
							Type:   corev1.NodeNetworkUnavailable,
							Status: corev1.ConditionFalse,
						},
						{
							Type:   corev1.NodeMemoryPressure,
							Status: corev1.ConditionFalse,
						},
						{
							Type:   corev1.NodeDiskPressure,
							Status: corev1.ConditionFalse,
						},
						{
							Type:   corev1.NodePIDPressure,
							Status: corev1.ConditionFalse,
						},
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					}
					err = clusterclient.Status().Update(ctx, node1)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns true", func(ctx SpecContext) {
					actual, err := healthCheckService.Do(ctx, "node-1")
					Expect(actual).To(BeTrue())
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when node is unschedulable", func() {
				BeforeEach(func(ctx SpecContext) {
					node2, err := clientset.CoreV1().Nodes().Apply(ctx, node2ApplyConfiguration, metav1.ApplyOptions{FieldManager: fieldManager})
					Expect(err).NotTo(HaveOccurred())

					node2.Status.Conditions = []corev1.NodeCondition{
						{
							Type:   corev1.NodeNetworkUnavailable,
							Status: corev1.ConditionFalse,
						},
						{
							Type:   corev1.NodeMemoryPressure,
							Status: corev1.ConditionFalse,
						},
						{
							Type:   corev1.NodeDiskPressure,
							Status: corev1.ConditionFalse,
						},
						{
							Type:   corev1.NodePIDPressure,
							Status: corev1.ConditionFalse,
						},
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					}
					err = clusterclient.Status().Update(ctx, node2)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns false", func(ctx SpecContext) {
					actual, err := healthCheckService.Do(ctx, "node-2")
					Expect(actual).To(BeFalse())
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
