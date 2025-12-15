# NovaStream Backend Makefile

.PHONY: build run test clean deps fmt lint

# Build the application
build:
	go build -o bin/novastream main.go

# Run the application
run:
	go run main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	go clean
	rm -rf bin/

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golint ./...

# Run all checks
check: fmt lint test

# Development server with hot reload (requires air)
dev:
	air

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/novastream-linux-amd64 ./main.go
	GOOS=windows GOARCH=amd64 go build -o bin/novastream-windows-amd64.exe ./main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/novastream-darwin-amd64 ./main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/novastream-darwin-arm64 ./main.go

# Docker build
docker-build:
	docker build -t novastream-backend .

# Docker run
docker-run:
	docker run -p 7777:7777 novastream-backend

# Install development tools
install-tools:
	go install github.com/cosmtrek/air@latest
	go install golang.org/x/lint/golint@latest
	go install github.com/go-delve/delve/cmd/dlv@latest


