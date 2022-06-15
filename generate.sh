#!/bin/bash
set -e

OUTPUT=pkg/clientset
rm -rf $OUTPUT

./bin/client-gen --go-header-file hack/boilerplate.go.txt \
  --clientset-name versioned \
  --input-base \
  --input engytita/v1alpha1 \
  --input-dirs github.com/engytita/engytita-operator/api/v1alpha1 \
  --trim-path-prefix=github.com/engytita/engytita-operator \
  --output-package github.com/engytita/engytita-operator/$OUTPUT \
  --output-base ./ \
  --clientset-api-path /api \
  -v 10

du -h $OUTPUT
cat $OUTPUT/**/*
tree $OUTPUT
