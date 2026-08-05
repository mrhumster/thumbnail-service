MODULE_NAME := github.com/mrhumster/thumbnail-service
PROTO_DIR := proto/stream
GEN_DIR := gen/go

.PHONY: build test proto vet

build:
	go build -o thumbnail-worker ./cmd/worker/main.go

test:
	go test -v ./...

vet:
	go vet ./...

proto:
	@echo "Generate protoc"
	mkdir -p $(GEN_DIR)
	protoc --proto_path=$(PROTO_DIR) \
		--go_out=. --go-grpc_out=. \
		--go_opt=module=$(MODULE_NAME) \
		--go-grpc_opt=module=$(MODULE_NAME) \
		$(PROTO_DIR)/*.proto
	@echo "Proto file generated in $(GEN_DIR)"
