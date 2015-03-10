.PHONY: all godep test build arm

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
