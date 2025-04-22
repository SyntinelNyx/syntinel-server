env ?= development
port ?= 8080

syntinel-server:
	@go build ./cmd/syntinel-server
run:
	@go run ./cmd/syntinel-server -e $(env) -p $(port)
test:
	@go test ./... -v
test_coverage:
	@go test -cover ./...
clean:
	@rm ./syntinel-server
proto:
	@protoc --go_out=. --go-grpc_out=. internal/proto/control.proto
