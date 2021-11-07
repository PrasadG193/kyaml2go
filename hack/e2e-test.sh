#!/bin/bash

set -o errexit
set -o nounset

make build

# Test core resources with typed client
for spec in ./testdata/*.yaml; do
  echo "testing $spec with typed client"
  # Typed client
  kyaml2go create < $spec > testdata/result.go
  go run testdata/result.go
  kyaml2go get < $spec > testdata/result.go
  go run testdata/result.go
  echo
  kyaml2go delete < $spec > testdata/result.go
  go run testdata/result.go
  echo "---------------------"
done

# Test CRs with typed client
# Create CRDs
for spec in ./testdata/crds/*.yaml; do
  echo "testing $spec with typed client"
  kyaml2go create < $spec > testdata/result.go
  go run testdata/result.go
  kyaml2go get < $spec > testdata/result.go
  go run testdata/result.go
  echo
  echo "---------------------"
done

# Test Foo CR
spec="./testdata/crs/foo.yaml"
echo "testing $spec with typed client"
kyaml2go create --cr --scheme "k8s.io/sample-controller/pkg/generated/clientset/versioned/scheme" --apis "k8s.io/sample-controller/pkg/apis/samplecontroller" --client "k8s.io/sample-controller/pkg/generated/clientset/versioned" < $spec > testdata/result.go
go run testdata/result.go
kyaml2go get --cr --scheme "k8s.io/sample-controller/pkg/generated/clientset/versioned/scheme" --apis "k8s.io/sample-controller/pkg/apis/samplecontroller" --client "k8s.io/sample-controller/pkg/generated/clientset/versioned" < $spec > testdata/result.go
go run testdata/result.go
echo
kyaml2go delete --cr --scheme "k8s.io/sample-controller/pkg/generated/clientset/versioned/scheme" --apis "k8s.io/sample-controller/pkg/apis/samplecontroller" --client "k8s.io/sample-controller/pkg/generated/clientset/versioned" < $spec > testdata/result.go
go run testdata/result.go

# Test Backup CR
spec="./testdata/crs/velero-backup.yaml"
echo "testing $spec"
kyaml2go create --cr --scheme "github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned/scheme" --apis "github.com/vmware-tanzu/velero/pkg/apis/velero" --client "github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned" < $spec > testdata/result.go
go run testdata/result.go
kyaml2go get --cr --scheme "github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned/scheme" --apis "github.com/vmware-tanzu/velero/pkg/apis/velero" --client "github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned" < $spec > testdata/result.go
go run testdata/result.go
echo
kyaml2go delete --cr --scheme "github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned/scheme" --apis "github.com/vmware-tanzu/velero/pkg/apis/velero" --client "github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned" < $spec > testdata/result.go
echo "---------------------"

# Delete CRDs
for spec in ./testdata/crds/*.yaml; do
  kyaml2go delete < $spec > testdata/result.go
  go run testdata/result.go
  echo "---------------------"
done

## Test core resources with dynamic client
for spec in ./testdata/*.yaml; do
  echo "testing $spec with dynamic client"
  # Dynamic client
  kyaml2go create --dynamic < $spec > testdata/result.go
  go run testdata/result.go
  kyaml2go get --dynamic < $spec > testdata/result.go
  go run testdata/result.go
  echo
  kyaml2go delete --dynamic < $spec > testdata/result.go
  go run testdata/result.go
  echo "---------------------"
done

# Test CRs with dynamic client
# Create CRDs
for spec in ./testdata/crds/*.yaml; do
  echo "testing $spec with dynamic client"
  kyaml2go create --dynamic < $spec > testdata/result.go
  go run testdata/result.go
  kyaml2go get --dynamic < $spec > testdata/result.go
  go run testdata/result.go
  echo
  echo "---------------------"
done
# Create CRs
for spec in ./testdata/crs/*.yaml; do
  echo "testing $spec with dynamic client"
  kyaml2go create --dynamic < $spec > testdata/result.go
  go run testdata/result.go
  kyaml2go get --dynamic < $spec > testdata/result.go
  go run testdata/result.go
  echo
  kyaml2go delete --dynamic < $spec > testdata/result.go
  go run testdata/result.go
  echo "---------------------"
done
# Delete CRDs
for spec in ./testdata/crds/*.yaml; do
  kyaml2go delete --dynamic < $spec > testdata/result.go
  go run testdata/result.go
  echo "---------------------"
done

rm testdata/result.go

echo "PASS"
