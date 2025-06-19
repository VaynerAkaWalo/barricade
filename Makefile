build:
	@go build -o bin/app cmd/main.go
	@chmod +x bin/app

run: build
	@./bin/app

.PHONY: build run