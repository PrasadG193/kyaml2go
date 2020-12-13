#!/bin/bash

set -o errexit
set -o nounset

make build

# Test core resources
for spec in ./testdata/*.yaml; do
  echo "testing $spec"
  kyaml2go create < $spec > testdata/result.go
  go run testdata/result.go
  kyaml2go get < $spec > testdata/result.go
  go run testdata/result.go
  echo
  kyaml2go delete < $spec > testdata/result.go
  go run testdata/result.go
  echo "---------------------"
done

# Test CRs
# Create CRDs
for spec in ./testdata/crds/*.yaml; do
  echo "testing $spec"
  kyaml2go create < $spec > testdata/result.go
  go run testdata/result.go
  kyaml2go get < $spec > testdata/result.go
  go run testdata/result.go
  echo
  echo "---------------------"
done

# Test CRs
# Create CRDs
# Test Foo CR
spec="./testdata/crs/foo.yaml"
echo "testing $spec"
kyaml2go create --cr --scheme "k8s.io/sample-controller/pkg/generated/clientset/versioned/scheme" --apis "k8s.io/sample-controller/pkg/apis/samplecontroller" --client "k8s.io/sample-controller/pkg/generated/clientset/versioned" < $spec > testdata/result.go
go run testdata/result.go
kyaml2go get --cr --scheme "k8s.io/sample-controller/pkg/generated/clientset/versioned/scheme" --apis "k8s.io/sample-controller/pkg/apis/samplecontroller" --client "k8s.io/sample-controller/pkg/generated/clientset/versioned" < $spec > testdata/result.go
go run testdata/result.go
echo
kyaml2go delete --cr --scheme "k8s.io/sample-controller/pkg/generated/clientset/versioned/scheme" --apis "k8s.io/sample-controller/pkg/apis/samplecontroller" --client "k8s.io/sample-controller/pkg/generated/clientset/versioned" < $spec > testdata/result.go
go run testdata/result.go
echo "---------------------"

# Delete CRDs
for spec in ./testdata/crds/*.yaml; do
  kyaml2go delete < $spec > testdata/result.go
  go run testdata/result.go
  echo "---------------------"
done

rm testdata/result.go

echo "PASS"
