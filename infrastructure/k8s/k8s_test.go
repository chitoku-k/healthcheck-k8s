package k8s_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestAction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8s Suite")
}

const (
	fieldManager = "healthcheck-k8s"
)

var (
	env           envtest.Environment
	clientset     *kubernetes.Clientset
	clusterclient client.Client
)

var _ = BeforeSuite(func(ctx SpecContext) {
	var err error
	config, err := env.Start()
	Expect(err).NotTo(HaveOccurred())

	manager, err := ctrl.NewManager(config, ctrl.Options{
		MetricsBindAddress: "0",
	})
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()

		err := manager.Start(context.Background())
		Expect(err).NotTo(HaveOccurred())
	}()

	Eventually(manager.Elected()).Should(BeClosed())
	clusterclient = manager.GetClient()

	clientset, err = kubernetes.NewForConfig(config)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := env.Stop()
	Expect(err).NotTo(HaveOccurred())
})
