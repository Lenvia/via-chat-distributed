package models

import (
	"context"
	"cvm/pb/gpt"
	"fmt"
	"log"
)

type GptMsgServer struct {
	gpt.UnimplementedGptMsgSenderServer
}

func (*GptMsgServer) Send(ctx context.Context, req *gpt.GptMsgRequest) (*gpt.GptMsgResponse, error) {
	query := req.GetQuery()

	fmt.Println("received query: ", query)

	reply, err := GetReply(OpenaiClient, query)
	if err != nil {
		log.Println(err)
		return &gpt.GptMsgResponse{Reply: "gpt connection timeout."}, err
	}
	res := &gpt.GptMsgResponse{Reply: reply}
	return res, nil
}
