#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="kwok-scale-poc"
KUBECONFIG_PATH=".kwok/kubeconfig"

if kwokctl get clusters 2>/dev/null | grep -qx "${CLUSTER_NAME}"; then
	echo "cluster ${CLUSTER_NAME} already exists; ensuring it is started"
	kwokctl start cluster --name="${CLUSTER_NAME}" >/dev/null 2>&1 || true
else
	kwokctl create cluster --name="${CLUSTER_NAME}" --runtime=binary
fi

mkdir -p "$(dirname "${KUBECONFIG_PATH}")"
kwokctl get kubeconfig --name="${CLUSTER_NAME}" > "${KUBECONFIG_PATH}"
echo "kubeconfig: ${KUBECONFIG_PATH}"
