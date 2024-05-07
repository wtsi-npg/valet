VERSION := $(shell git describe --always --tags --dirty)
ldflags := "-X github.com/wtsi-npg/valet/valet.Version=${VERSION}"
build_path = "build/valet-${VERSION}"

.PHONY: build coverage dist install lint test check clean

all: build

install:
	go install -ldflags ${ldflags}

build:
	mkdir -p ${build_path}
	go build -v -ldflags ${ldflags} -o ${build_path}/valet github.com/wtsi-npg/valet

lint:
	golangci-lint run ./...

check: test

test:
	ginkgo -r --race

coverage:
	ginkgo -r --cover -coverprofile=coverage.out

dist: build test
	cp README.md COPYING ${build_path}
	mkdir ${build_path}/scripts
	cp scripts/valet_archive_create.sh ${build_path}/scripts/
	tar -C ./build -cvj -f valet-${VERSION}.tar.bz2 valet-${VERSION}
	shasum -a 256 valet-${VERSION}.tar.bz2 > valet-${VERSION}.tar.bz2.sha256

clean:
	go clean
	rm -rf build/*
