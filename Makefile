.DEFAULT_GOAL := build
.PHONY: build pre-build system-check cli

#Build the binary
build: pre-build
	@cd cmd/exec;GOOS_VAL=$(shell go env GOOS) GOARCH_VAL=$(shell go env GOARCH) go build -o $(shell go env GOPATH)/bin/kyaml2go
	@cd cmd/cli;GOOS_VAL=$(shell go env GOOS) GOARCH_VAL=$(shell go env GOARCH) go build -o $(shell go env GOPATH)/bin/kyaml2go_cli
	@cd cmd/serve;GOOS_VAL=$(shell go env GOOS) GOARCH_VAL=$(shell go env GOARCH) go build -o $(shell go env GOPATH)/bin/kyaml2go_serve
	@echo "Build completed successfully"

cli: 
	@cd cmd/cli;GOOS_VAL=$(shell go env GOOS) GOARCH_VAL=$(shell go env GOARCH) go build -o $(shell go env GOPATH)/bin/kyaml2go_cli
	@echo "cli generated successfully"

#system checks
system-check:
	@echo "Checking system information"
	@if [ -z "$(shell go env GOOS)" ] || [ -z "$(shell go env GOARCH)" ] ; \
	then \
	echo 'ERROR: Could not determine the system architecture.' && exit 1 ; \
	else \
	echo 'GOOS: $(shell go env GOOS)' ; \
	echo 'GOARCH: $(shell go env GOARCH)' ; \
	echo 'System information checks passed.'; \
	fi ;

#Pre-build checks
pre-build: system-check
