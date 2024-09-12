env ?= development
port ?= 0

run:
	@go run ./cmd/syntinel -e $(env) -p $(port) 
build:
	@go build ./cmd/syntinel
test:
	@go test ./...
clean:
	@rm ./syntinel