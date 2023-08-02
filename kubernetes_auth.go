//go:build !authless

package main

// Enable auth plugins from client-go such as "oidc".
import _ "k8s.io/client-go/plugin/pkg/client/auth"
