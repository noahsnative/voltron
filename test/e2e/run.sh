#!/bin/bash

set -e

scriptsLocation="$(dirname "${0}")"

function install_kind() 
{
  curl -Lo ./kind "https://github.com/kubernetes-sigs/kind/releases/download/v0.7.0/kind-$(uname)-amd64"
  chmod +x ./kind
}

function cleanup()
{
  kind delete cluster
  rm -rf ./kind
}

trap cleanup EXIT

install_kind

kind create cluster --image "kindest/node:v1.17.0"

source "${scriptsLocation}"/webhook-create-signed-cert.sh --namespace default --service webhook --secret webhook-cert

exit 0

