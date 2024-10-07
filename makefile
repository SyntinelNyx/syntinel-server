env ?= development
port ?= 0

syntinel-server:
	@go build ./cmd/syntinel-server
run:
	@go run ./cmd/syntinel-server -e $(env) -p $(port) 	
test:
	@go test ./...
clean:
	@rm ./syntinel-server