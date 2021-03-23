package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
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

	server := grpc.NewServer()
	pb.RegisterHelloServer(server, &Server{})

	wrappedGrpc := grpcweb.WrapServer(server, grpcweb.WithWebsockets(true))

	httpServer := http.Server{
		ErrorLog: log.Default(),
		Addr:     ":4000",
		Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			resp.Header().Set("Access-Control-Allow-Origin", "*")
			resp.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			resp.Header().Set("Access-Control-Allow-Headers", "*")

			if req.Method == "OPTIONS" {
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

	log.Println("starting to serve at :4000")

	httpServer.ListenAndServeTLS("./certs/cert.pem", "./certs/privkey.pem")
}
