.PHONY: build run test clean deps docker

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=sentinel

# Build the application
build: build-web
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/server

# Build frontend
build-web:
	cd web && npm install && npm run build

# Run the application
run:
	$(GORUN) ./cmd/server

# Run frontend dev server
dev-web:
	cd web && npm run dev

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf ./data/*.db
	rm -rf ./web/dist

# Build Docker image
docker:
	docker build -t code-sentinel:latest .

# Run with Docker Compose
docker-up:
	docker-compose up -d

# Stop Docker Compose
docker-down:
	docker-compose down

# View logs
logs:
	docker-compose logs -f

# Format code
fmt:
	$(GOCMD) fmt ./...

# Lint code
lint:
	golangci-lint run ./...

# Initialize project (first time setup)
init: deps
	mkdir -p data configs
	cp configs/config.example.yaml configs/config.yaml
	@echo "Please edit configs/config.yaml with your settings"
