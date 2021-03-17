gen_types: hello.proto
	protoc --go_out=types --go-grpc_out=types hello.proto

run: gen_types
	go run .

build: gen_types
	go build -o server .

all: build
