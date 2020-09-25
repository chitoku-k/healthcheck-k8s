healthcheck-k8s
===============

[![][workflow-badge]][workflow-link]

Check if the specified Kubernetes node is schedulable and return as HTTP status
code.

## Requirements

- Go
- Kubernetes

## Installation

```sh
$ go build
```

```sh
# Port number (required)
export PORT=8080

# Name of header in which client sends a node name (required)
export HEADER_NAME=X-Node

# Path to the kubeconfig, or else falls back to service account token mounted inside the Pod (optional)
export KUBECONFIG=$HOME/.kube/config

# Timeout in milliseconds (optional; zero means infinity)
export TIMEOUT_MS=30000
```

## Usage

```sh
$ ./healthcheck-k8s
```

### Normal: node is schedulable

```sh
$ curl --dump-header - -H 'X-Node: minikube' localhost:8080
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Date: Wed, 01 Jan 2020 00:00:00 GMT
Content-Length: 26

Node(s) are OK: "minikube"
```

### Cordoned: node is unschedulable

```sh
$ kubectl cordon minikube
node/minikube uncordoned

$ curl --dump-header - -H 'X-Node: minikube' localhost:8080
HTTP/1.1 503 Service Unavailable
Content-Type: text/plain; charset=utf-8
Date: Wed, 01 Jan 2020 00:00:00 GMT
Content-Length: 52

Node "minikube" is currently undergoing maintenance.
```

### Spec

| Status | Condition                                     |
|--------|-----------------------------------------------|
| 200    | Node is schedulable.                          |
| 400    | Header is not present in the request.         |
| 404    | Node was not found.                           |
| 500    | Unexpected error when retrieving node status. |
| 503    | Node is unschedulable.                        |
| 504    | Timed out connecting to kube-apiserver.       |

[workflow-link]:    https://github.com/chitoku-k/healthcheck-k8s/actions?query=branch:master
[workflow-badge]:   https://img.shields.io/github/workflow/status/chitoku-k/healthcheck-k8s/CI%20Workflow/master.svg?style=flat-square
