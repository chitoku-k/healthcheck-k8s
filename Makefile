SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: setup-envtest
setup-envtest:
	GOBIN=$(shell pwd)/bin go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
