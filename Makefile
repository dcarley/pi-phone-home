.PHONY: all godep test build arm docker

all: test build

godep:
	go get github.com/tools/godep
	godep restore

test: godep
	godep go test ./...

build: godep
	godep go build

arm: export GOARCH = arm
arm: export GOARM = 7
arm: build

docker:
	docker build --force-rm -qt pi-phone-home .
	docker run --rm -ti \
		-v ${GOPATH}:/gopath \
		-v ${GOPATH}/bin/linux_amd64:/gopath/bin \
		-w /gopath/src/$(shell go list) \
		snappy
