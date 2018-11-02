FROM alpine:3.6

RUN apk update && apk add ca-certificates

USER nobody

ADD build/_output/bin/k8s-controller-datadog-monitor /usr/local/bin/k8s-controller-datadog-monitor
