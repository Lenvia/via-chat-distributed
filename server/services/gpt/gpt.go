package gpt

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	gogpt "github.com/sashabaranov/go-openai"
	"gopkg.in/ini.v1"
	"log"
	"net/http"
	"net/url"
	"via-chat/models"
)

var OpenaiClient *openai.Client

var ChatGptName string

func LoadGPT(file *ini.File) {
	ApiKey := file.Section("gpt").Key("API_KEY").String()
	ChatGptName = file.Section("gpt").Key("GPT_NAME").String()

	if len(ApiKey) < 10 {
		return
	}

	config := gogpt.DefaultConfig(ApiKey)
	proxyUrl, err := url.Parse("http://127.0.0.1:7890")
	//proxyUrl, err := url.Parse("http://host.docker.internal:7890") // 访问宿主机代理
	if err != nil {
		log.Println(err)
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}
	config.HTTPClient = &http.Client{
		Transport: transport,
	}

	OpenaiClient = gogpt.NewClientWithConfig(config)

	if OpenaiClient != nil {
		user := models.FindUserByField("username", ChatGptName)
		if user.ID <= 0 {
			_ = models.AddUser(map[string]interface{}{
				"username":  ChatGptName,
				"password":  "NULL",
				"avatar_id": "1",
			})
		}
	}
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
