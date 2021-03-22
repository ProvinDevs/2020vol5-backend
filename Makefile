TYPE_FILES=types/hello.pb.go types/hello_grpc.pb.go

$(TYPE_FILES): hello.proto
	protoc --go_out=types --go-grpc_out=types hello.proto

fmt: $(TYPE_FILES)
	go fmt ./...

run: $(TYPE_FILES) fmt
	go run -v .

lint: $(TYPE_FILES) fmt
	go vet ./...

test: $(TYPE_FILES) fmt
	go test ./...

server.a: $(TYPE_FILES) fmt
	go build -v -o server.a .

all: sever.a 
