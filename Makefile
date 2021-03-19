gen_types: hello.proto
	protoc --go_out=types --go-grpc_out=types hello.proto

fmt: gen_types
	go fmt ./...

run: gen_types fmt
	go run -v .

build: gen_types fmt
	go build -v -o server.a .

lint: gen_types fmt
	go vet ./...

test: gen_types fmt
	go test ./...

all: build
