package cvm

import (
	"cvm/models"
	"cvm/pb/gpt"
	"fmt"
	"google.golang.org/grpc"
	"net"
)

func main() {
	models.InitGPT()

	l, _ := net.Listen("tcp", ":8888")
	s := grpc.NewServer()
	gpt.RegisterGptMsgSenderServer(s, &models.GptMsgServer{})
	s.Serve(l)
	fmt.Println("listening...")
}
