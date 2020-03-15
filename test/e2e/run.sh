#!/bin/bash

set -e
trap 'kind delete cluster' EXIT

curl -Lo ./kind "https://github.com/kubernetes-sigs/kind/releases/download/v0.7.0/kind-$(uname)-amd64"
chmod +x ./kind
kind create cluster --image "kindest/node:v1.17.0"
sh ${BASH_SOURCE[0]%/*}/webhook-create-signed-cert.sh --namespace default --service webhook --secret webhook-cert
exit 0




