#!/bin/bash
set -e

OUTPUT=pkg/clientset
rm -rf $OUTPUT

./bin/client-gen --go-header-file hack/boilerplate.go.txt \
  --clientset-name versioned \
  --input-base '' \
  --input github.com/engytita/engytita-operator/pkg/apis/cache/v1alpha1 \
  --trim-path-prefix=github.com/engytita/engytita-operator \
  --output-package github.com/engytita/engytita-operator/$OUTPUT \
  --output-base ./ \
  -v 10

du -h $OUTPUT
cat $OUTPUT/**/*
tree $OUTPUT
