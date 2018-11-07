# Build the manager binary
FROM golang:1.10.3 as builder

# Copy in the go src
WORKDIR /go/src/github.com/j-vizcaino/k8s-controller-datadog-monitor
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/j-vizcaino/k8s-controller-datadog-monitor/cmd/manager

# Copy the controller-manager into a thin image
FROM alpine:3.7

RUN apk update && apk add ca-certificates

WORKDIR /usr/local/bin
COPY --from=builder /go/src/github.com/j-vizcaino/k8s-controller-datadog-monitor/manager .

USER nobody

ENTRYPOINT ["./manager"]
