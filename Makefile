.PHONY: build run test clean deps lint fmt build-frontend dev check-frontend

APP_NAME := grape
VERSION := 0.1.0
BUILD_DIR := ./bin

# 检查前端资源完整性
check-frontend:
	@echo "Checking frontend assets..."
	@if [ ! -f "internal/web/dist/index.html" ]; then \
		echo "❌ Error: Frontend assets not found!"; \
		echo "   Please run 'make build-frontend' first or use 'make build'"; \
		exit 1; \
	fi
	@if [ ! -d "internal/web/dist/assets" ]; then \
		echo "❌ Error: Frontend assets directory missing!"; \
		echo "   Please run 'make build-frontend' first"; \
		exit 1; \
	fi
	@echo "✅ Frontend assets OK"

build-frontend:
	@echo "Building frontend..."
	cd web && npm run build
	@echo "Copying frontend to embed directory..."
	rm -rf internal/web/dist
	cp -r web/dist internal/web/

build: build-frontend check-frontend
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/grape

build-only: check-frontend
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

# 开发模式 - 构建后运行（Air 热重载 Go 代码）
# 需要安装 air: go install github.com/air-verse/air@latest
AIR_BIN := $(shell go env GOPATH)/bin/air

dev:
	@echo "========================================="
	@echo "  Grape Development Mode"
	@echo "========================================="
	@echo "Building frontend..."
	@cd web && npm run build
	@echo "Copying frontend to embed directory..."
	@rm -rf internal/web/dist
	@cp -r web/dist internal/web/
	@echo "========================================="
	@echo "Starting backend with Air (hot reload)"
	@echo "Web UI:      http://localhost:4873"
	@echo "API Port:    http://localhost:4874"
	@echo "========================================="
	@echo "Modify Go code → Air auto-reloads"
	@echo "Modify Vue code → Re-run make dev"
	@echo "========================================="
	@if [ ! -f "$(AIR_BIN)" ]; then \
		echo "Installing air..."; \
		go install github.com/air-verse/air@latest; \
	fi
	$(AIR_BIN) -c .air.toml