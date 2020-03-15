#!/bin/bash

set -e
trap 'kind delete cluster' EXIT

go get -u sigs.k8s.io/kind
kind create cluster --image "kindest/node:v1.17.0"
sh ${BASH_SOURCE[0]%/*}/webhook-create-signed-cert.sh --namespace default --service webhook --secret webhook-cert
exit 0




