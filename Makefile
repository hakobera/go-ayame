VERSION ?= 0.1.0

.PHONY: test version

all: test

test:
	go test ./...

version:
	@echo $(VERSION)