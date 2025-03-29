package config_test

import (
	"os"
	"time"

	"github.com/chitoku-k/healthcheck-k8s/infrastructure/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get()", func() {
	BeforeEach(func() {
		err := os.Unsetenv("PORT")
		Expect(err).NotTo(HaveOccurred())

		err = os.Unsetenv("HEADER_NAME")
		Expect(err).NotTo(HaveOccurred())

		err = os.Unsetenv("TIMEOUT_MS")
		Expect(err).NotTo(HaveOccurred())

		err = os.Unsetenv("TRUSTED_PROXIES")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when configuration is invalid", func() {
		Context("when environment variables are missing", func() {
			It("returns an error", func() {
				_, err := config.Get()
				Expect(err).To(MatchError(And(
					HavePrefix("missing env(s):"),
					ContainSubstring("PORT"),
					ContainSubstring("HEADER_NAME"),
				)))
			})
		})

		Context("when timeout cannot be parsed", func() {
			BeforeEach(func() {
				err := os.Setenv("PORT", "8080")
				Expect(err).NotTo(HaveOccurred())

				err = os.Setenv("HEADER_NAME", "X-Node")
				Expect(err).NotTo(HaveOccurred())

				err = os.Setenv("TIMEOUT_MS", "1000ms")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := config.Get()
				Expect(err).To(MatchError(HavePrefix("timeout is invalid:")))
			})
		})
	})

	Context("when configuration is valid", func() {
		BeforeEach(func() {
			err := os.Setenv("PORT", "8080")
			Expect(err).NotTo(HaveOccurred())

			err = os.Setenv("HEADER_NAME", "X-Node")
			Expect(err).NotTo(HaveOccurred())

			err = os.Setenv("TIMEOUT_MS", "5000")
			Expect(err).NotTo(HaveOccurred())

			err = os.Setenv("TRUSTED_PROXIES", "127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns environment", func() {
			actual, err := config.Get()
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(config.Environment{
				Port:       "8080",
				HeaderName: "X-Node",
				Timeout:    5000 * time.Millisecond,
				TrustedProxies: []string{
					"127.0.0.0/8",
					"10.0.0.0/8",
					"172.16.0.0/12",
					"192.168.0.0/16",
				},
			}))
		})
	})
})
