.PHONY: all godep test build arm snappy snappy-build snappy-deploy docker

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

snappy: snappy-build snappy-deploy

snappy-build:
	@if [ -z $${PHONE_URL} ]; then \
		echo "PHONE_URL must be set"; \
		exit 1; \
	fi
	sed "s|<URL>|$${PHONE_URL}|" meta/package.yaml.tmpl > meta/package.yaml
	snappy build .

snappy-deploy:
	@if [ -z $${SNAPPY_URL} ]; then \
		echo "SNAPPY_URL must be set"; \
		exit 1; \
	fi
	snappy-remote --url $${SNAPPY_URL} install $(shell ls -1 *.snap | tail -n1)

docker:
	docker build --force-rm -qt pi-phone-home .
	docker run --rm -ti \
		-v ${GOPATH}:/gopath \
		-v ${GOPATH}/bin/linux_amd64:/gopath/bin \
		-w /gopath/src/$(shell go list) \
		snappy
