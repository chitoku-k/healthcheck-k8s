//go:build windows

package k8s_test

import (
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = AfterSuite(func() {
	err := env.Stop()
	Expect(err).To(MatchError(syscall.Errno(syscall.EWINDOWS)))
})
