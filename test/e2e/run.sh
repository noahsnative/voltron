#!/bin/bash

set -e

scriptLocation="$(dirname "${0}")"

function install_kind() 
{
  curl -Lo ./kind "https://github.com/kubernetes-sigs/kind/releases/download/v0.7.0/kind-$(uname)-amd64"
  chmod +x ./kind
}

function create_kind_cluster()
{
  reg_name='kind-registry'
  reg_port='5000'
  running="$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)"
  if [ "${running}" != 'true' ]; then
    docker run \
      -d --restart=always -p "${reg_port}:5000" --name "${reg_name}" \
      registry:2
  fi
  reg_ip="$(docker inspect -f '{{.NetworkSettings.IPAddress}}' "${reg_name}")"

  echo "Retagging the image and pushing to the local registry"
  docker tag voltron/injector:latest localhost:5000/voltron/injector:latest
  docker push localhost:5000/voltron/injector:latest

  cat <<EOF | kind create cluster --image "kindest/node:v1.17.0"  --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches: 
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
    endpoint = ["http://${reg_ip}:${reg_port}"]
EOF
}

function cleanup()
{
  kind delete cluster
  rm -rf ./kind
}

trap cleanup EXIT

install_kind

create_kind_cluster

source "${scriptLocation}"/webhook-create-signed-cert.sh --namespace default --service webhook --secret webhook-cert

kubectl apply -f "${scriptLocation}"/pod.yaml

echo "waiting for pod to start"
for x in $(seq 10); do
  podReadyStatus=$(kubectl get pods -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}')
  if [[ ${podReadyStatus} == 'True' ]]; then
      break
  fi
  sleep 30
done
if [[ ${podReadyStatus} != 'True' ]]; then
  echo "ERROR: pod is not ready. Giving up after 10 attempts." >&2
  exit 1
fi

kubectl get pods

exit 0

