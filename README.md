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

and put your certs to `certs/{cert.pem,privkey.pem}`.
grpc-web needs to run with this proxy and also needs these certs for TLS.

## launch

Just run `make run`. It requires that certs file (cert.pem, privkey.pem) in /certs/ folder

After that, you can access on port 4000.
