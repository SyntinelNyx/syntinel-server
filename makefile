run:
	@go run ./cmd/syntinel
build:
	@go build ./cmd/syntinel
test:
	@go test ./...
clean:
	@rm ./syntinel