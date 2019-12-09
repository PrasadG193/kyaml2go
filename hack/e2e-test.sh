#!/bin/bash

set -o errexit
set -o nounset

make build

for spec in ./testdata/*.yaml; do
  echo "testing $spec"
  kyaml2go create -f $spec > testdata/result.go
  go run testdata/result.go
  kyaml2go get -f $spec > testdata/result.go
  go run testdata/result.go
  echo
  kyaml2go delete -f $spec > testdata/result.go
  go run testdata/result.go
  echo "---------------------"
done

rm testdata/result.go

echo "PASS"
