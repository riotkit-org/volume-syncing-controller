#!/usr/bin/env bash

# Copyright 2019 The Tekton Authors

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]:-$0}"; )" &> /dev/null && pwd 2> /dev/null; )";

# Conveniently set GOPATH if unset
if [[ -z "${GOPATH:-}" ]]; then
  export GOPATH="$(go env GOPATH)"
  if [[ -z "${GOPATH}" ]]; then
    echo "WARNING: GOPATH not set and go binary unable to provide it"
  fi
fi

# Useful environment variables
readonly REPO_ROOT_DIR="${REPO_ROOT_DIR:-$(git rev-parse --show-toplevel 2> /dev/null)}"
readonly REPO_NAME="${REPO_NAME:-$(basename ${REPO_ROOT_DIR} 2> /dev/null)}"

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
# This generates deepcopy,client,informer and lister for the resource package (v1alpha1)
# This is separate from the pipeline package as resource are staying in v1alpha1 and they
# need to be separated (at least in terms of go package) from the pipeline's packages to
# not having dependency cycle.
bash ${REPO_ROOT_DIR}/hack/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/riotkit-org/volume-syncing-controller/pkg/client github.com/riotkit-org/volume-syncing-controller/pkg/apis \
  "riotkit.org:v1alpha1" \
  --go-header-file="${SCRIPT_DIR}/boilerplate.go.txt" \
  --output-base=${SCRIPT_DIR}/../.build/generated

# clean up
rm -rf ${SCRIPT_DIR}/../pkg/client

# copy regenerated
cp -pr ${SCRIPT_DIR}/../.build/generated/github.com/riotkit-org/volume-syncing-controller/pkg/apis/* ${SCRIPT_DIR}/../pkg/apis
mv ${SCRIPT_DIR}/../.build/generated/github.com/riotkit-org/volume-syncing-controller/pkg/client ${SCRIPT_DIR}/../pkg/client
rm -rf ${SCRIPT_DIR}/../.build/generated
