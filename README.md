## make all

You need to setup `protoc` to build.

#### macOS

```
brew install protobuf
```

#### ArchLinux

```
sudo pacman -S protobuf
```

then run this:

```
go get -v google.golang.org/protobuf/cmd/protoc-gen-go google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

## make proxy

You need to install `grpcwebproxy`.

```
go get -v github.com/improbable-eng/grpc-web/go/grpcwebproxy
```

and put your certs to `certs/{cert.pem,privkey.pem}`.
grpc-web needs to run with this proxy and also needs these certs for TLS.

## launch

Just run `make run`, then run `make proxy` on another terminal.

After that, you can access on port 3000.
