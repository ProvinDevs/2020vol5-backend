gen_types: hello.proto
	protoc --go_out=types --go-grpc_out=types hello.proto

run: gen_types
	go run -v .

build: gen_types
	go build -v -o server .

all: build
