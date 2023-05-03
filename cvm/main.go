package main

import (
	"cvm/models"
	"cvm/pb/gpt"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	models.InitGPT()

	s := grpc.NewServer()
	gpt.RegisterGptMsgSenderServer(s, &models.GptMsgServer{})
	fmt.Println("rpc 注册完成！")

	l, _ := net.Listen("tcp", ":8765")
	err := s.Serve(l)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("listening...")
}
