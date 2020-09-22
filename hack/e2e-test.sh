#!/bin/bash

set -o errexit
set -o nounset

make build

for spec in ./testdata1/*.yaml; do
  echo "testing $spec"
  kyaml2go create -f $spec > testdata1/result.go
  go run testdata1/result.go
  kyaml2go get -f $spec > testdata1/result.go
  go run testdata1/result.go
  echo
  kyaml2go delete -f $spec > testdata1/result.go
  go run testdata1/result.go
  echo "---------------------"
done

rm testdata1/result.go

echo "PASS"
