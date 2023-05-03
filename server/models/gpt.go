package models

import (
	"fmt"
	"google.golang.org/grpc"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"via-chat-distributed/pb/gpt"
)

var GptClient gpt.GptMsgSenderClient
var ChatGptName string

func CreateBot(file *ini.File) {
	ChatGptName = file.Section("gpt").Key("GPT_NAME").String()
	user := FindUserByField("username", ChatGptName)
	if user.ID <= 0 {
		_ = AddUser(map[string]interface{}{
			"username":  ChatGptName,
			"password":  "NULL",
			"avatar_id": "1",
		})
	}

}

func InitGptClient() {
	l, err := grpc.Dial("127.0.0.1:8765", grpc.WithInsecure())
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

	fmt.Println("Gpt is ready.\n", GptClient)

}
