# ISV CLI VERSION
CLI_VERSION ?= 0.4.0
CLI_PLATFORM ?= linux
CLI_ARCH ?= amd64
CLI_IMG ?= quay.io/jooholee/isv-cli:${CLI_VERSION}

.PHONY: build
build: test 
	sed "s/cliVersion =.*/cliVersion = \"$(CLI_VERSION)\"/g" -i ./pkg/cli/cli.go
	go build ./cmd/isv-cli.go
	cp isv-cli ./build/.
	# ./hack/build.sh
test:
	go test ./cmd/isv-cli.go

cli-image: podman-build podman-push

podman-build: download
	sed "s/cliVersion =.*/cliVersion = \"$(CLI_VERSION)\"/g" -i ./pkg/cli/cli.go
	podman build -t ${CLI_IMG} -f $(shell pwd)/build/Dockerfile.isv-cli .

podman-push:
	podman push ${CLI_IMG}

download:
	test -f ./build/isv-cli || wget https://github.com/Jooho/isv-cli/releases/download/${CLI_VERSION}/isv-cli_${CLI_PLATFORM}_${CLI_ARCH}
	test -f ./build/isv-cli || mv ./isv-cli_${CLI_PLATFORM}_${CLI_ARCH} ./build/isv-cli

clean:
	rm ./isv-cli ./build/isv-cli
