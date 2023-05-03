package primary

import (
	"github.com/gin-gonic/gin"
	"via-chat-distributed/ws"
	"via-chat-distributed/ws/go_ws"
)

func Create() ws.ServeInterface {
	return &go_ws.GoServe{}
}

// Start 启动 websocket
func Start(gin *gin.Context) {
	// 根据配置文件中，`app.serve_type` 键中对应的值创建 serve 实例，并启动服务
	Create().RunWs(gin)
}

// OnlineUserCount 返回在线用户的数量
func OnlineUserCount() int {
	return Create().GetOnlineUserCount()
}

// OnlineRoomUserCount 返回指定房间在线用户的数量
func OnlineRoomUserCount(roomId int) int {
	return Create().GetOnlineRoomUserCount(roomId)
}
