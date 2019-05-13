export GOPATH := $(shell pwd)
export PATH := $(GOPATH)/bin:$(PATH)

all: get build test

get:
	@echo "--> go get relate package..."
	go get github.com/xelabs/go-mysqlstack/driver
	go get github.com/stretchr/testify/assert
	go get github.com/pierrre/gotestcover

build:
	@$(MAKE) get
	@echo "--> Building godumper ..."
	go build -v -o bin/dumper src/mydumper/main.go
	go build -v -o bin/loader src/myloader/main.go
	go build -v -o bin/streamer src/mystreamer/main.go
	@chmod 755 bin/*

clean:
	@echo "--> Cleaning..."
	@go clean
	@rm -f bin/*

fmt:
	go fmt ./...
	go vet ./...

test:
	@echo "--> Testing..."
	@$(MAKE) testcommon

testcommon:
	go test -race -v common

# code coverage
COVPKGS =	common
coverage:
	go build -v -o bin/gotestcover \
	src/github.com/pierrre/gotestcover/*.go;
	gotestcover -coverprofile=coverage.out -v $(COVPKGS)
	go tool cover -html=coverage.out
.PHONY: get build clean fmt test coverage
