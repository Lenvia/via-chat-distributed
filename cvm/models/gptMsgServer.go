package models

import (
	"context"
	"cvm/pb/gpt"

	"log"
)

type GptMsgServer struct {
	gpt.UnimplementedGptMsgSenderServer
}

func (*GptMsgServer) Send(ctx context.Context, req *gpt.GptMsgRequest) (*gpt.GptMsgResponse, error) {
	query := req.GetQuery()

	reply, err := GetReply(OpenaiClient, query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	res := &gpt.GptMsgResponse{Reply: reply}
	return res, nil
}
