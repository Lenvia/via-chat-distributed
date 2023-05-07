package models

import (
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
	"log"
)

var NC *nats.Conn
var BaseTopic string

func InitMessageQueue() {
	url := viper.GetString("nats.ip") + ":" + viper.GetString("nats.port")
	BaseTopic = viper.GetString("base_topic")
	var err error
	NC, err = nats.Connect(url)
	if err != nil {
		log.Fatal(err)
	}
}
