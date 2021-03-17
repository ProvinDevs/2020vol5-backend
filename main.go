package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	pb "github.com/ProvinDevs/2020vol5-backend/types"
)

type Server struct {
	pb.UnimplementedHelloServer
}

func (s *Server) Greeting(_ context.Context, req *pb.GreetingRequest) (*pb.GreetingReply, error) {
	name := req.GetName()
	msg := fmt.Sprintf("Hello %s!", name)

	log.Println(msg)

	return &pb.GreetingReply{Message: msg}, nil
}

func main() {
	log.SetOutput(os.Stdout)
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	listener, err := net.Listen("tcp", ":"+port)

	if err != nil {
		log.Panicf("failed to listen :%s %v\n", port, err)
	}

	server := grpc.NewServer()
	pb.RegisterHelloServer(server, &Server{})

	log.Printf("starting to serve at :%s\n", port)

	err = server.Serve(listener)

	if err != nil {
		log.Panicf("failed to serve: %v", err)
	}
}
