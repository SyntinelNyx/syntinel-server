env ?= development
port ?= 8080

syntinel-server:
	@go build ./cmd/syntinel-server
run:
	@go run ./cmd/syntinel-server -e $(env) -p $(port)
test:
	@go test ./... -v
clean:
	@rm ./syntinel-server
proto:
	protoc --go_out=./ --go_opt=paths=source_relative \
	--go-grpc_out=./ --go-grpc_opt=paths=source_relative \
	./internal/proto/hardware_info.proto
