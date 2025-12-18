sqlc:
	@echo "Sqlc generate"
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@cd config && sqlc generate

swag:
	@go run github.com/swaggo/swag/cmd/swag@latest init -g internal/adapter/in/http/router.go --output docs/ --parseDependency --parseInternal

# Generating code
gen: sqlc swag
	@echo "Generating code..."
	echo "Code generated successfully"

# ===== Dependencies =====
deps: gen
	@echo "Installing dependencies..."
	@go generate ./...
	@go mod download
	@go mod tidy

# ===== Build =====
ARTIFACT_VERSION ?= 0.0.0-local

build: gen deps
	@echo "Building version $(ARTIFACT_VERSION)..."
	@mkdir -p bin
	@go build \
		-o ./bin/auth \
		./cmd/auth

# ===== Run =====
run: build
	@echo "Running binary..."
	./bin/auth

# ===== Lint =====
lint:
	@echo "Linting..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run --tests=false --disable-all --timeout=2m -p error


# ===== Tests =====
test: mocks
	@echo "Testing..."
	@go test -v -race ./internal/...

coverage: mocks
	@echo "Coverage..."
	@go test -race -coverprofile=coverage.out -covermode=atomic ./internal/...
	@go tool cover -html=coverage.out -o coverage.html

# ===== Docker image =====
build-image:
	@echo "Building docker image version $(ARTIFACT_VERSION)..."
	@docker build \
		--build-arg ARTIFACT_VERSION=$(ARTIFACT_VERSION) \
		-t enduran-auth:$(ARTIFACT_VERSION) .

# ===== Cleanup =====
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf docs/
	@rm -f coverage.out coverage.html

# ===== Help (default) =====
help:
	@echo "Available commands:"
	@echo "  gen          - Generate sqlc + swagger"
	@echo "  d
