# syntax = docker/dockerfile:1
FROM golang:1.20.7 AS build
WORKDIR /usr/src
COPY go.mod go.sum /usr/src/
RUN --mount=type=cache,target=/go \
    go mod download
COPY . /usr/src/
ARG TAGS
ARG VERSION=v0.0.0-dev
RUN --mount=type=cache,target=/go \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -tags="$TAGS" -ldflags="-s -w -X main.version=$VERSION"

FROM scratch
ENV GIN_MODE release
ENV PORT 80
COPY --from=build /usr/src/healthcheck-k8s /healthcheck-k8s
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 80
CMD ["/healthcheck-k8s"]
