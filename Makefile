.PHONY: build run test clean deps lint fmt build-frontend

APP_NAME := grape
VERSION := 0.1.0
BUILD_DIR := ./bin

build-frontend:
	@echo "Building frontend..."
	cd web && npm run build
	@echo "Copying frontend to embed directory..."
	rm -rf internal/web/dist
	cp -r web/dist internal/web/

build: build-frontend
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/grape

build-only:
	@echo "Building $(APP_NAME) (without frontend)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/grape

run:
	go run ./cmd/grape

run-with-config:
	go run ./cmd/grape -c ./configs/config.yaml

test:
	go test -v ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf ./data
	@rm -rf ./web/dist

deps:
	go mod tidy
	go mod download

lint:
	golangci-lint run

fmt:
	go fmt ./...

dev:
	@echo "Starting development environment..."
	@echo "Backend: http://localhost:4873"
	@echo "Frontend: http://localhost:3000"
	cd web && npm run dev &
	go run ./cmd/grape