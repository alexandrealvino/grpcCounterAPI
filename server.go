package main

import (
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"grpcCounterAPI/chat"
	"net"
	"sync"
)

func main() {
	log.Info("Server is running!")
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal("Failed to listen on port 9000: ", err, lis)
	}
	s := chat.Server{A: 0, Mutex: &sync.Mutex{}}
	grpcServer := grpc.NewServer()
	chat.RegisterChatServiceServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to listen on port 9000: ", err)
	}
}