.PHONY: generate build run dev test clean

generate:
	go generate ./...

build: generate
	go build -o bin/recipebox .

run: build
	./bin/recipebox serve

dev:
	$(shell go env GOPATH)/bin/air

test:
	go test ./...

clean:
	rm -rf bin/
