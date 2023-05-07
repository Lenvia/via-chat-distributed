package models

import "time"

var SMsg = make(chan WebSocketMsg) // 发送的消息，用于处理客户端的消息

// MsgData 结构体定义了消息体的数据结构
type MsgData struct {
	Uid      string        `json:"uid"`       // 发送者 uid
	Username string        `json:"username"`  // 发送者用户名
	AvatarId string        `json:"avatar_id"` // 发送者头像 id
	ToUid    string        `json:"to_uid"`    // 接收者 uid
	Content  string        `json:"content"`   // 消息内容
	ImageUrl string        `json:"image_url"` // 图片地址
	RoomId   string        `json:"room_id"`   // 房间 id
	Count    int           `json:"count"`     // 房间人数
	List     []interface{} // 房间中其他客户端信息
	Time     int64         // 消息发送时间
	// 下面是数据库额外附加信息，兼容一下
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// WebSocketMsg 结构体定义了 WebSocket 消息体
type WebSocketMsg struct {
	Status int     `json:"status"` // 消息状态码
	Data   MsgData `json:"data"`   // 消息体数据
	//Conn   *websocket.Conn // 对应的客户端连接对象
}
