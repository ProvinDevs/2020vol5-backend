package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	pb "github.com/ProvinDevs/2020vol5-backend/types"
)

func main() {
	rand.Seed(time.Now().Unix())
	log.SetOutput(os.Stdout)
	port := os.Getenv("PORT")

	if port == "" {
		port = "4000"
	}

	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalln("PORT must be valid port number")
	}

	port = ":" + port

	server := grpc.NewServer()
	pb.RegisterHelloServer(server, &Server{})

	wrappedGrpc := grpcweb.WrapServer(server, grpcweb.WithWebsockets(true))

	httpServer := http.Server{
		ErrorLog: log.Default(),
		Addr:     port,
		Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			resp.Header().Set("Access-Control-Allow-Origin", "*")
			resp.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			resp.Header().Set("Access-Control-Allow-Headers", "*")

			if req.Method == "OPTIONS" {
				return
			}

			if req.URL.Path == "/.well-known/signalling/health" {
				return
			}

			if wrappedGrpc.IsGrpcWebRequest(req) {
				wrappedGrpc.ServeHTTP(resp, req)
			}

			if wrappedGrpc.IsGrpcWebSocketRequest(req) {
				wrappedGrpc.HandleGrpcWebsocketRequest(resp, req)
			}
		}),
	}

	var err error
	noTLS := os.Getenv("NO_TLS")

	if noTLS != "" {
		log.Printf("starting to serve at %s without TLS\n", port)
		err = httpServer.ListenAndServe()
	} else {
		log.Printf("starting to serve at %s\n", port)
		err = httpServer.ListenAndServeTLS("./certs/cert.pem", "./certs/privkey.pem")
	}

	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
