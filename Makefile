.PHONY: generate build run test clean

generate:
	go generate ./...

build: generate
	go build -o bin/recipebox .

run: build
	./bin/recipebox serve

test:
	go test ./...

clean:
	rm -rf bin/
