package test

import (
	"github.com/gorilla/websocket"
	"log"
	"net/url"
)

func StartFunc(strI string) {
	var addr = "localhost:8322"

	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// 进入房间
	d := make(map[string]interface{})
	d["status"] = 1
	d["data"] = map[string]interface{}{
		"uid":       strI,
		"room_id":   "1",
		"avatar_id": "1",
		"username":  "random_" + strI,
	}
	c.WriteJSON(d)

	// 说话
	d["status"] = 3
	d["data"] = map[string]interface{}{
		"uid":       strI,
		"room_id":   "1",
		"avatar_id": "1",
		"username":  "random_" + strI,
		"content":   "hello" + strI,
		"to_uid":    "0",
	}

	err = c.WriteJSON(d)
	if err != nil {
		log.Println(err)
	}

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
	}

}
