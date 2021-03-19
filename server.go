package main

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/ProvinDevs/2020vol5-backend/types"

	"github.com/google/uuid"
)

type Room struct {
	id uuid.UUID
}

type Server struct {
	pb.UnimplementedHelloServer

	mu    sync.Mutex
	rooms []*Room
}

func (s *Server) CreateRoom(_ context.Context, _ *pb.CreateRoomRequest) (*pb.CreateRoomPayload, error) {
	newRoom := Room{
		id: uuid.New(),
	}

	s.mu.Lock()
	s.rooms = append(s.rooms, &newRoom)
	s.mu.Unlock()

	return &pb.CreateRoomPayload{RoomId: newRoom.id.String()}, nil
}

func (s *Server) Signalling(_ pb.Hello_SignallingServer) error {
	return fmt.Errorf("unimplemented")
}
