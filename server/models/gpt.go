package models

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"log"
	"via-chat-distributed/pb/gpt"
)

var GptClient gpt.GptMsgSenderClient
var ChatGptName string
var ChatGptIdInt int

func CreateBot() {
	ChatGptName = viper.GetString("gpt.gpt_name")
	user := FindUserByField("username", ChatGptName)
	if user.ID <= 0 {
		user = AddUser(User{
			Username: ChatGptName,
			Password: "NULL",
			AvatarId: "1",
		})

	}
	ChatGptIdInt = int(user.ID)
}

func InitGptClient() {
	url := viper.GetString("cvm.ip") + ":" + viper.GetString("cvm.port")
	l, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		log.Println(err)
		return
	}
	GptClient = gpt.NewGptMsgSenderClient(l)

	// gpt
	CreateBot()

	log.Println("Gpt is ready.")

}
