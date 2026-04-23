.PHONY: test run tidy build

tidy:
	go mod tidy

test:
	go test ./...

run:
	go run ./cmd/worker

build:
	go build ./cmd/worker
