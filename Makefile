.PHONY: run build test cover fmt vet check

run:
	go run ./cmd/api

build:
	go build ./...

test:
	go test ./...

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

fmt:
	gofmt -w .

vet:
	go vet ./...

check: fmt vet test build
