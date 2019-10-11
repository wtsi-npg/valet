VERSION := $(shell git describe --always --tags --dirty)
ldflags := "-X valet/valet.Version=${VERSION}"
build_path = "build/valet-${VERSION}"

.PHONY: build coverage dist install lint test check clean

all: build

install:
	go install -ldflags ${ldflags}

build:
	mkdir -p ${build_path}
	go build -v -ldflags ${ldflags} -o ${build_path}/valet github.com/kjsanger/valet

lint:
	golangci-lint run ./...

check: test

test:
	ginkgo -r -slowSpecThreshold=60 -race

coverage:
	ginkgo -r -slowSpecThreshold=60 -cover -coverprofile=coverage.out

dist: build lint test build
	cp README.md COPYING ${build_path}
	tar -C ./build -cvj -f valet-${VERSION}.tar.bz2 valet-${VERSION}
	shasum -a 256 valet-${VERSION}.tar.bz2 > valet-${VERSION}.tar.bz2.sha256

clean:
	go clean
	rm -rf build/*
