.PHONY: build test install clean

build:
	go build -ldflags="-s -w" -o luma .

test:
	go test ./...

install:
	go install .

clean:
	rm -f luma
