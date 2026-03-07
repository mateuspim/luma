.PHONY: build test install clean

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o luma .

test:
	go test ./...

install:
	go install .

clean:
	rm -f luma
