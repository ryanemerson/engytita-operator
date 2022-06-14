#!/bin/bash
set -e

module_name="github.com/cert-manager/cert-manager"
client_subpackage="pkg/client"
client_package="${module_name}/${client_subpackage}"
# Generate clientsets, listers and informers for user-facing API types
client_inputs=(
  pkg/apis/certmanager/v1 \
  pkg/apis/acme/v1 \
)

# Generate defaulting functions to be used by the mutating webhook
defaulter_inputs=(
  internal/apis/certmanager/v1alpha2 \
  internal/apis/certmanager/v1alpha3 \
  internal/apis/certmanager/v1beta1 \
  internal/apis/certmanager/v1 \
  internal/apis/acme/v1alpha2 \
  internal/apis/acme/v1alpha3 \
  internal/apis/acme/v1beta1 \
  internal/apis/acme/v1 \
  internal/apis/config/webhook/v1alpha1 \
  internal/apis/meta/v1 \
  pkg/webhook/handlers/testdata/apis/testgroup/v2 \
  pkg/webhook/handlers/testdata/apis/testgroup/v1 \
)

# Generate conversion functions to be used by the conversion webhook
conversion_inputs=(
  internal/apis/certmanager/v1alpha2 \
  internal/apis/certmanager/v1alpha3 \
  internal/apis/certmanager/v1beta1 \
  internal/apis/certmanager/v1 \
  internal/apis/acme/v1alpha2 \
  internal/apis/acme/v1alpha3 \
  internal/apis/acme/v1beta1 \
  internal/apis/acme/v1 \
  internal/apis/config/webhook/v1alpha1 \
  internal/apis/meta/v1 \
  pkg/webhook/handlers/testdata/apis/testgroup/v2 \
  pkg/webhook/handlers/testdata/apis/testgroup/v1 \
)

gen-clientsets() {
  echo "${client_subpackage}"/clientset
  echo "Generating clientset..." >&2
  prefixed_inputs=( "${client_inputs[@]/#/$module_name/}" )
  joined=$( IFS=$','; echo "${prefixed_inputs[*]}" )

  echo client-gen \
    --go-header-file hack/boilerplate/boilerplate.generatego.txt \
    --clientset-name versioned \
    --input-base "" \
    --input "$joined" \
    --trim-path-prefix="$module_name" \
    --output-package "${client_package}"/clientset \
    --output-base ./
}

gen-clientsets
