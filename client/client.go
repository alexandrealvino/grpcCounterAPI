package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"grpcCounterAPI/chat"
	"strconv"
	"sync"
	"time"
)

func main() {
	begin := time.Now()
	goRoutines := 40
	var wg sync.WaitGroup
	out := make([]string, goRoutines + 1)
	for i := 1; i < len(out); i++ {
		wg.Add(1)
		go CreateClient(i, &out[i], &wg, goRoutines)
	}
	wg.Wait()
	println(time.Since(begin).String())
}

func CreateClient(i int, outVal *string, wg *sync.WaitGroup, goRoutines int)  {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":9000", grpc.WithInsecure())
	if err != nil {
		log.Fatal("Could not connect: ", err)
	}
	defer conn.Close()

	c := chat.NewChatServiceClient(conn)
	message := chat.Message{Body: "Hello from client", A: strconv.Itoa(i)}
	now := time.Now().Minute()
	for now < 1 {
		time.Sleep(1000 * time.Millisecond)
		now = time.Now().Minute()
		fmt.Print(now)
	}
	limit := 0
	for limit <= 10000 - goRoutines {
		response, err := c.SayHello(context.Background(), &message)
		if err != nil {
			log.Fatal("Error when calling server: ", err)
		}
		log.Info(response.Body, " count: ", response.A)
		time.Sleep(time.Millisecond)
		limit,_ = strconv.Atoi(response.A)
	}
	*outVal = ""
	wg.Done()
}
