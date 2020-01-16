NAME=csi-zfs-plugin
OS ?= linux
VERSION ?= test
all: test

publish: compile build

compile:
	@echo "==> Building the project"
	CGO_ENABLED=0 go build --ldflags "-extldflags -static" -mod=vendor  -o ./cmd/csi-zfs/csi-zfs-plugin ./cmd/csi-zfs/main.go

test:
	@echo "==> Testing all packages"
	@go test -v ./...

build:
	@echo "==> Building the docker image"
	@docker build -t duni/csi-zfs-plugin:$(VERSION) cmd/csi-zfs -f cmd/csi-zfs/Dockerfile

vendor:
	@GO111MODULE=on go mod tidy
	@GO111MODULE=on go mod vendor

.PHONY: vendor build test compile