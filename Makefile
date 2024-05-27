SHELL_PATH = /bin/zsh
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/zsh,/bin/bash)

GOTEST_PACKAGES?=$(shell find . -name '*_test.go' -exec dirname {} \; | grep -v generated | uniq)
GOTEST_FLAGS=-race

.PHONY: deps
deps:
	@go mod tidy && go mod vendor && go mod verify

.PHONE: test
test:
	go test $(GOTEST_FLAGS) -v $(GOTEST_PACKAGES)