#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

go vet ./pkg/...
go vet ./cmd/...
