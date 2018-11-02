#!/usr/bin/env bash
set -ex

tag=${1-latest}
operator-sdk build quay.io/jerome.vizcaino/k8s-controller-datadog-monitor:$tag
docker push quay.io/jerome.vizcaino/k8s-controller-datadog-monitor:$tag

