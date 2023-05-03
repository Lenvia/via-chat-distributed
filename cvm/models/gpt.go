package models

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	gogpt "github.com/sashabaranov/go-openai"
	"gopkg.in/ini.v1"
	"log"
	"os"
)

var OpenaiClient *openai.Client

func InitGPT() {
	gptConfigFilePath := "configs/openai_config.ini"
	// 使用 os.Stat() 函数获取文件的状态信息
	_, err := os.Stat(gptConfigFilePath)
	if err == nil {
		// 文件存在
		file, err := ini.Load("configs/openai_config.ini")
		if err != nil {
			fmt.Println("配置文件读取错误:", err)
		}
		LoadGPT(file)
		log.Println("初始化GPT完成!")
	} else {
		log.Println(err)
	}
}

func LoadGPT(file *ini.File) {
	ApiKey := file.Section("gpt").Key("API_KEY").String()
	if len(ApiKey) < 10 {
		return
	}

	config := gogpt.DefaultConfig(ApiKey)
	OpenaiClient = gogpt.NewClientWithConfig(config)

}

func GetReply(client *openai.Client, query string) (string, error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: query},
			},
		},
	)
	if err != nil {
		return "", err
	}

	log.Println(resp)
	var reply string
	if len(resp.Choices) > 0 {
		reply = resp.Choices[0].Message.Content
	} else {
		reply = ""
		err = errors.Errorf("failed")
	}

	return reply, err
}
