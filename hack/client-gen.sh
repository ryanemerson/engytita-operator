#!/bin/bash
set -e

PROJECT_ROOT=$1
CLIENT_GEN=$2
OUTPUT_PACKAGE=$3

PKG_ROOT="github.com/engytita/engytita-operator"
APIS_PKG="${PKG_ROOT}/pkg/apis"
APIS_DIR="${PROJECT_ROOT}/pkg/apis"

rm -rf "${OUTPUT_PACKAGE}"

# client-gen only seems to play nice with source files in /pkg/apis/<kind>/<version>, so temporarily create structure
mkdir -p "${APIS_DIR}"/cache/v1alpha1 "${APIS_DIR}"/cacheregion/v1alpha1
cp "$PROJECT_ROOT"/api/v1alpha1/cache_types.go "${APIS_DIR}"/cache/v1alpha1/
cp "$PROJECT_ROOT"/api/v1alpha1/cacheregion_types.go "${APIS_DIR}"/cacheregion/v1alpha1/

# Ensure that the types compile
sed -i "s#SchemeBuilder#//SchemeBuilder#g" "${APIS_DIR}"/cache/v1alpha1/cache_types.go "${APIS_DIR}"/cacheregion/v1alpha1/cacheregion_types.go

"${CLIENT_GEN}" --go-header-file hack/boilerplate.go.txt \
  --clientset-name versioned \
  --input-base '' \
  --input "${APIS_PKG}"/cache/v1alpha1,"${APIS_PKG}"/cacheregion/v1alpha1 \
  --trim-path-prefix=${PKG_ROOT} \
  --output-package ${PKG_ROOT}/"${OUTPUT_PACKAGE}" \
  --output-base ./ \
  -v 10

# Update the clientset imports to use the actual types
find "${OUTPUT_PACKAGE}" -type f | xargs sed -i "s#${PKG_ROOT}/pkg/apis/cache/v1alpha1#${PKG_ROOT}/api/v1alpha1#g"
find "${OUTPUT_PACKAGE}" -type f | xargs sed -i "s#${PKG_ROOT}/pkg/apis/cacheregion/v1alpha1#${PKG_ROOT}/api/v1alpha1#g"

# Clean up tmp apis dir
#rm -rf "${APIS_DIR}"
