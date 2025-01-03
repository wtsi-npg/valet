VERSION := $(shell git describe --always --tags --dirty)
ldflags := "-X github.com/wtsi-npg/valet/valet.Version=${VERSION}"
build_args := -a -v -ldflags ${ldflags}
build_path = "build/valet-${VERSION}"

CGO_ENABLED := 1
GOARCH := amd64
GOOS := linux

ifeq ($(GITHUB_ACTIONS),true)
DOCKER_REGISTRY?=ghcr.io
DOCKER_USER?=$(GITHUB_REPOSITORY_OWNER)
else
DOCKER_REGISTRY?=docker.io
DOCKER_USER?=wsinpg
endif
TAG=${VERSION}

NOW=$(shell date --utc --iso-8601=seconds)

DOCKER_PREFIX?=$(DOCKER_REGISTRY)/$(DOCKER_USER)
DOCKER_ARGS?=--platform linux/amd64 --progress=plain --rm

image_names := valet
images := $(addsuffix .$(TAG), $(image_names))
remote := $(addsuffix .$(TAG).pushed, $(image_names))

git_url=$(shell git remote get-url origin)
git_commit=$(shell git log --pretty=format:'%H' -n 1)

.PHONY: build check clean coverage dist docker install lint push test

all: build

build:
	mkdir -p ${build_path}
	go build ${build_args} -o ${build_path}/valet-${GOARCH} ./main.go

install:
	go install ${build_args}

lint:
	golangci-lint run ./...

check: test

test:
	ginkgo -r --race

coverage:
	ginkgo -r --cover -coverprofile=coverage.out

dist: build
	cp README.md COPYING ${build_path}
	mkdir -p ${build_path}/scripts
	cp scripts/valet_archive_create.sh ${build_path}/scripts/
	tar -C ./build -cvj -f ./build/valet-${VERSION}.tar.bz2 valet-${VERSION}
	shasum -a 256 ./build/valet-${VERSION}.tar.bz2 > ./build/valet-${VERSION}.tar.bz2.sha256

docker: $(images)

valet.$(TAG): Dockerfile
	docker buildx build ${DOCKER_ARGS} \
    --label org.opencontainers.image.title="valet ${VERSION} Linux ${GOARCH}" \
    --label org.opencontainers.image.source=$(git_url) \
    --label org.opencontainers.image.revision=$(git_commit) \
    --label org.opencontainers.image.version=$(TAG) \
    --label org.opencontainers.image.created=$(NOW) \
    --tag $(DOCKER_PREFIX)/valet:$(VERSION) \
    --tag $(DOCKER_PREFIX)/valet:latest \
    --file $< .
	touch $@

push: $(remote)

%.$(TAG).pushed: %.$(TAG)
	echo docker push $(DOCKER_PREFIX)/$*:$(TAG)
	echo docker push $(DOCKER_PREFIX)/$*:latest
	touch $@

clean:
	go clean
	$(RM) -r ./build
	rm -f $(foreach image_name,$(image_names), $(image_name).*)
