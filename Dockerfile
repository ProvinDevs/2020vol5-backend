FROM golang:alpine3.13 as build

RUN apk add --no-cache protobuf make && \
    go get -v google.golang.org/protobuf/cmd/protoc-gen-go google.golang.org/grpc/cmd/protoc-gen-go-grpc && \
    mkdir -p /src/types

COPY go.mod go.sum hello.proto main.go server.go Makefile /src/
WORKDIR /src
RUN make all


FROM alpine:3.13
COPY --from=build /src/server.a /usr/local/bin/server

CMD ["/usr/local/bin/server"]

