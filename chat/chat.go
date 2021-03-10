package chat

import (
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"strconv"
	"sync"
)

type Server struct {
	A int
	Mutex *sync.Mutex
}

func (s *Server) SayHello(_ context.Context, message *Message) (*Message, error){
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.A++
	log.Info(message.Body, " ", message.A)
	return &Message{Body: "Hello from server!", A: strconv.Itoa(s.A)}, nil
}

//generate proto
//protoc --go_out=plugins=grpc:chat chat.proto