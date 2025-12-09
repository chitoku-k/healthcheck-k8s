//go:build unix

package k8s_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = AfterSuite(func() {
	err := env.Stop()
	Expect(err).NotTo(HaveOccurred())
})
