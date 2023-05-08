package models

import (
	"fmt"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"via-chat-distributed/pb/gpt"
)

var GptClient gpt.GptMsgSenderClient
var ChatGptName string
var ChatGptIdInt int

func CreateBot(file *ini.File) {
	ChatGptName = file.Section("gpt").Key("GPT_NAME").String()
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
	gptConfigFilePath := "configs/openai_config.ini"
	// 使用 os.Stat() 函数获取文件的状态信息
	_, err = os.Stat(gptConfigFilePath)
	if err == nil {
		// 文件存在
		file, err := ini.Load(gptConfigFilePath)
		if err != nil {
			fmt.Println("配置文件读取错误:", err)
		}
		CreateBot(file)
	} else {
		log.Println(err)
	}

	log.Println("Gpt is ready.")

}
