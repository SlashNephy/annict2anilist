build: build-batch build-authorize

build-batch:
	go build -o batch ./cmd/batch

build-authorize:
	go build -o authorize ./cmd/authorize

run-batch:
	go run ./cmd/batch

run-authorize:
	go run ./cmd/authorize

test:
	go test ./...
