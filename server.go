package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"sync"

	pb "github.com/ProvinDevs/2020vol5-backend/types"
)

type Room struct {
	id            int32
	joinedUserIds map[string]pb.Hello_SignallingServer
}

const (
	roomIdMin    = 10000
	roomIdMax    = 99999
	roomMaxCount = roomIdMax - roomIdMin
)

type Server struct {
	pb.UnimplementedHelloServer

	mu    sync.Mutex
	rooms []*Room
}

func (s *Server) newRoom() *Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.rooms) >= roomMaxCount {
		log.Fatalln("roomMaxCount exceeded")
	}

retry:
	id := int32((rand.Int() % (roomIdMax - roomIdMin + 1)) + roomIdMin)

	for _, v := range s.rooms {
		if v.id == id {
			goto retry
		}
	}

	newRoom := Room{id: id, joinedUserIds: make(map[string]pb.Hello_SignallingServer)}
	s.rooms = append(s.rooms, &newRoom)

	log.Printf("New Room(%d) has created\n", newRoom.id)
	printCurrentRooms(&s.rooms)

	return &newRoom
}

func (s *Server) CreateRoom(_ context.Context, _ *pb.CreateRoomRequest) (*pb.Room, error) {
	newRoom := s.newRoom()

	pbRoom := pb.Room{
		RoomId:        int32(newRoom.id),
		JoinedUserIds: []string{},
	}

	return &pbRoom, nil
}

func (s *Server) Signalling(stream pb.Hello_SignallingServer) error {
	worker := Worker{
		mu:     &s.mu,
		rooms:  &s.rooms,
		room:   nil,
		st:     stream,
		userId: "",
	}

	return worker.start()
}

type Worker struct {
	mu     *sync.Mutex
	rooms  *[]*Room
	room   *Room
	st     pb.Hello_SignallingServer
	userId string
}

func (w *Worker) start() error {
	return w.recvRoutine()
}

func (w *Worker) recvRoutine() error {
	for {
		msg, err := w.st.Recv()

		if err == io.EOF {
			w.onStreamClose()
			return nil
		}

		if err != nil {
			w.onStreamClose()
			return err
		}

		w.onMessage(msg)
	}
}

func (w *Worker) onStreamClose() {
	log.Printf("User %s has left\n", w.userId)

	w.mu.Lock()
	defer w.mu.Unlock()

	emptyRoomIndexes := []int{}

	for index, room := range *w.rooms {
		delete(room.joinedUserIds, w.userId)

		if len(room.joinedUserIds) == 0 {
			emptyRoomIndexes = append(emptyRoomIndexes, index)
		}
	}

	// remove empty rooms from w.rooms
	for _, index := range emptyRoomIndexes {
		log.Printf("Room %d has dropped.\n", (*w.rooms)[index].id)

		(*w.rooms)[index] = (*w.rooms)[len(*w.rooms)-1]
		(*w.rooms)[len(*w.rooms)-1] = &Room{}
		*w.rooms = (*w.rooms)[:len(*w.rooms)-1]
	}

	printCurrentRooms(w.rooms)
}

func (w *Worker) onMessage(msg *pb.SendSignallingMessage) {
	body := msg.GetBody()

	switch typedBody := body.(type) {
	case *pb.SendSignallingMessage_SelfIntro:
		log.Println("SelfIntroMessage has came")
		w.onSelfIntroduce(typedBody.SelfIntro)

	case *pb.SendSignallingMessage_RoomInfoRequest:
		log.Printf("RoomInfoRequest has came from %s\n", w.userId)
		w.onRoomInfoRequest()

	case *pb.SendSignallingMessage_SdpMessage:
		log.Printf("SdpMessage has came from %s\n", w.userId)
		w.onSdpMessage(typedBody.SdpMessage)

	case *pb.SendSignallingMessage_IceMessage:
		log.Printf("IceMessage has came from %s\n", w.userId)
		w.onIceCandidateMessage(typedBody.IceMessage)

	default:
		log.Printf("%s has sent message which has unknown message in body: %#v\n", w.userId, body)
	}
}

func (w *Worker) sendSelfIntroResult(ok bool, msg string) {
	w.st.Send(
		&pb.RecvSignallingMessage{
			Body: &pb.RecvSignallingMessage_SelfIntroResult{
				SelfIntroResult: &pb.SelfIntroduceResult{
					Ok:           ok,
					ErrorMessage: msg,
				},
			},
		},
	)
}

func (w *Worker) onSelfIntroduce(msg *pb.SelfIntroduceMessage) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.userId != "" && w.room != nil {
		log.Printf("%s has self-introduced twice\n", w.userId)
		w.sendSelfIntroResult(false, "this is your second self-introducing and it's not allowed.")

		return
	}

	userId := msg.GetMyId()
	roomId := msg.GetRoomId()

	w.userId = userId

	ok := false

	for _, v := range *w.rooms {
		if _, exists := v.joinedUserIds[userId]; exists {
			log.Printf("Duplicated UserID %s\n", userId)
			w.sendSelfIntroResult(false, "UserID is duplicated. Change to another one.")
		}
	}

	for _, v := range *w.rooms {
		if v.id == roomId {
			w.room = v
			w.room.joinedUserIds[userId] = w.st
			ok = true
			break
		}
	}

	if !ok {
		log.Printf("%s tried to join room %d which doesn't exist.\n", userId, roomId)
		w.sendSelfIntroResult(false, "specified room doesn't exist")

		return
	}

	w.sendSelfIntroResult(true, "")
	log.Printf("User %s joined to room %d\n", msg.GetMyId(), roomId)
}

func (w *Worker) onRoomInfoRequest() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.room == nil {
		log.Println("? has requested RoomInfo without self-introducing")
		return
	}

	joinedUserIds := make([]string, 0, len(w.room.joinedUserIds))

	for k := range w.room.joinedUserIds {
		joinedUserIds = append(joinedUserIds, k)
	}

	msg := pb.RecvSignallingMessage{
		Body: &pb.RecvSignallingMessage_RoomInfoResponse{
			RoomInfoResponse: &pb.Room{
				RoomId:        w.room.id,
				JoinedUserIds: joinedUserIds,
			},
		},
	}

	w.st.Send(&msg)
}

func (w *Worker) onSdpMessage(msg *pb.SendSdpMessage) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.userId == "" {
		log.Printf("? has requested to send Sdp to %s without self-introducing\n", msg.GetToId())
		return
	}

	toId := msg.GetToId()

	sendMsg := &pb.RecvSignallingMessage{
		Body: &pb.RecvSignallingMessage_SdpMessage{
			SdpMessage: &pb.RecvSdpMessage{
				SessionDescription: msg.GetSessionDescription(),
				FromId:             w.userId,
				ToId:               toId,
			},
		},
	}

	w.sendToOtherUserInSameRoom(toId, sendMsg)
}

func (w *Worker) onIceCandidateMessage(msg *pb.SendIceCandidateMessage) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.userId == "" {
		log.Printf("? has requested to send IceCandidate to %s without self-introducing\n", msg.GetToId())
		return
	}

	toId := msg.GetToId()

	sendMsg := &pb.RecvSignallingMessage{
		Body: &pb.RecvSignallingMessage_IceMessage{
			IceMessage: &pb.RecvIceCandidateMessage{
				IceCandidate: msg.GetIceCandidate(),
				FromId:       w.userId,
				ToId:         toId,
			},
		},
	}

	w.sendToOtherUserInSameRoom(toId, sendMsg)
}

// THIS FUNCTION EXPECT THAT w.mu IS ALREADY LOCKED
func (w *Worker) sendToOtherUserInSameRoom(userId string, msg *pb.RecvSignallingMessage) bool {
	for streamUserId, stream := range w.room.joinedUserIds {
		if streamUserId == userId {
			stream.Send(msg)
			return true
		}
	}

	return false
}

func printCurrentRooms(rooms *[]*Room) {
	buffer := ""
	for _, room := range *rooms {
		buffer += fmt.Sprintf("%d, ", room.id)
	}

	log.Printf("current rooms are: %s", buffer)
}
