group "default" {
    targets = ["healthcheck-k8s"]
}

target "healthcheck-k8s" {
    context = "."
}
