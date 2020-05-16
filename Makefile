VERSION ?= 0.2.0

EXAMPLE_DIRS := $(shell find ./examples/* -maxdepth 0 -type d)

.PHONY: all test version build $(EXAMPLE_DIRS)

all: test build

test:
	go test ./...

build: $(EXAMPLE_DIRS)

$(EXAMPLE_DIRS):
	cd $@ && go mod tidy && go build -o tmp/cmd .

version:
	@echo v$(VERSION)