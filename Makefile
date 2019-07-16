VERSION := $(shell git describe --always --tags --dirty)
ldflags := "-X valet/valet.Version=${VERSION}"

.PHONY: build clean install lint test

all: build

install:
	go install -ldflags ${ldflags}

build:
	mkdir -p ./build
	go build -ldflags ${ldflags} -o ./build/valet

lint:
	golangci-lint run ./...

test:
	go test -v ./...

clean:
	go clean
	rm -f build/valet
