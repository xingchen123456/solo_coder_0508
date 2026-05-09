.PHONY: all build clean proto run grpc http test

BINARY_NAME=game_backend
BUILD_DIR=./bin

all: build

proto:
	@echo "Generating proto files..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/game/*.proto

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)

run:
	@echo "Running all servers..."
	go run ./cmd/$(BINARY_NAME)

grpc:
	@echo "Running gRPC server..."
	go run ./cmd/$(BINARY_NAME) grpc

http:
	@echo "Running HTTP gateway server..."
	go run ./cmd/$(BINARY_NAME) http

test:
	@echo "Running tests..."
	go test -v ./...

tidy:
	@echo "Tidying dependencies..."
	go mod tidy
