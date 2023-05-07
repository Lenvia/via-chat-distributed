package models

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
	"log"
)

var NC *nats.Conn
var BaseTopic string

func InitMessageQueue() {
	url := viper.GetString("nats.ip") + ":" + viper.GetString("nats.port")
	BaseTopic = viper.GetString("nats.base_topic")
	var err error
	NC, err = nats.Connect(url)
	if err != nil {
		log.Println(err)
	}

	// 订阅消息
	_, err = NC.Subscribe(BaseTopic, func(message *nats.Msg) {
		chatMsg := WebSocketMsg{}
		err := json.Unmarshal(message.Data, &chatMsg)
		if err != nil {
			log.Println(err)
			return
		}
		SMsg <- chatMsg
	})

	if err != nil {
		log.Println(err)
	}
}
