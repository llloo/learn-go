.PHONY: build test bench vet lint run clean docker-build

BINARY=server
BUILD_DIR=build

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) cmd/server/main.go

test:
	go test -v ./...

bench:
	go test -bench=. -benchmem ./...

vet:
	go vet ./...

lint:
	$(shell go env GOPATH)/bin/golangci-lint run ./...

run:
	go run cmd/server/main.go

docker-build:
	docker build -t taskapi .

clean:
	rm -rf $(BUILD_DIR)
