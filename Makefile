# ISV CLI VERSION
CLI_VERSION ?= v0.1
CLI_PLATFORM ?= linux
CLI_ARCH ?= amd64
CLI_IMG ?= quay.io/jooholee/isv-cli:${CLI_VERSION}

build: 
	go build ./cmd/isv-cli.go

test:
	go test

cli-image: podman-build podman-push

podman-build: download
	podman build -t ${CLI_IMG} -f $(shell pwd)/build/Dockerfile.isv-cli .

podman-push:
	podman push ${CLI_IMG}

download:
	test -f ./build/isv-cli || wget https://github.com/Jooho/isv-cli/releases/download/${CLI_VERSION}/isv-cli_${CLI_PLATFORM}_${CLI_ARCH}
	test -f ./build/isv-cli || mv ./isv-cli_${CLI_PLATFORM}_${CLI_ARCH} ./build/isv-cli

clean:
	rm ./isv-cli
