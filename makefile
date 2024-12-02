env ?= development
port ?= 8080

syntinel-server:
	@go build ./cmd/syntinel-server
run:
	@go run ./cmd/syntinel-server -e $(env) -p $(port) 	
test:
	@go test ./...
clean:
	@rm ./syntinel-server

build_proto:
	protoc --go_out=./internal/gRPC_server/ --go_opt=paths=source_relative \
    --go-grpc_out=./internal/gRPC_server/  --go-grpc_opt=paths=source_relative \
    ./internal/gRPC_server/hw_info.proto