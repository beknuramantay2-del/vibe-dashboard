.PHONY: build dev test clean lint

# Wails build (production)
build:
	wails build

# Wails dev mode with hot-reload
dev:
	wails dev

# Run Go tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf build/bin/
	rm -f *.log

# Go vet + lint
lint:
	go vet ./...

# Install Wails CLI if not present
install-wails:
	go install github.com/wailsapp/wails/v2/cmd/wails@latest
