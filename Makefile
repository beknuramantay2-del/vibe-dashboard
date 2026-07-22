.PHONY: build run test clean install

BINARY = vibe-dashboard

build:
	go build -o $(BINARY) ./cmd/$(BINARY)

run: build
	./$(BINARY)

test:
	go test ./...

clean:
	rm -f $(BINARY)
	rm -f $(BINARY).exe
	rm -f *.log

install:
	go install ./cmd/$(BINARY)

lint:
	go vet ./...

cross:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux-amd64 ./cmd/$(BINARY)
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY)-darwin-amd64 ./cmd/$(BINARY)
	GOOS=windows GOARCH=amd64 go build -o $(BINARY)-windows-amd64.exe ./cmd/$(BINARY)
