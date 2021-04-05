#! /usr/bin/make -f

# Project variables.
PACKAGE := github.com/hermeznetwork/hermez-integration
VERSION := $(shell git describe --tags --always)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)
PROJECT_NAME := $(shell basename "$(PWD)")

# Go related variables.
GO_FILES ?= $$(find . -name '*.go' | grep -v vendor)
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin
GOPKG := $(.)
GOENVVARS := GOBIN=$(GOBIN)
GOCMD := $(GOBASE)
GOBINARY := hermez-integration

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECT_NAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## gofmt: Run `go fmt` for all go files.
gofmt:
	@echo "  >  Format all go files"
	$(GOENVVARS) gofmt -w ${GO_FILES}

## govet: Run go vet.
govet:
	@echo "  >  Running go vet"
	$(GOENVVARS) go vet ./...

## golint: Run default golint.
golint:
	@echo "  >  Running golint"
	$(GOENVVARS) golint -set_exit_status ./...

## exec: Run given command. e.g; make exec run="go test ./..."
exec:
	GOBIN=$(GOBIN) $(run)

## clean: Clean build files. Runs `go clean` internally.
clean:
	@-rm $(GOBIN)/ 2> /dev/null
	@echo "  >  Cleaning build cache"
	$(GOENVVARS) go clean

## build: Build the project.
build: install
	@echo "  >  Building the integration example binary..."
	$(GOENVVARS) go build -o $(GOBIN)/$(GOBINARY) $(GOCMD)

## install: Install missing dependencies. Runs `go get` internally. e.g; make install get=github.com/foo/bar
install:
	@echo "  >  Checking if there is any missing dependencies..."
	$(GOENVVARS) go get $(GOCMD)/... $(get)

## run: Run example.
run:
	@bash -c "$(MAKE) clean build"
	@echo "  >  Running $(PROJECT_NAME)"
	@$(GOBIN)/$(GOBINARY)
