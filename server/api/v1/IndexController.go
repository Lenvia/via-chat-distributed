package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"via-chat-distributed/services/helper"
	"via-chat-distributed/services/message_service"
	"via-chat-distributed/services/user_service"
	"via-chat-distributed/ws/primary"
)

// Index 函数用于显示应用程序的登录界面。
// 在用户已登录的情况下，将自动跳转到应用程序的房间页面。
// 在用户未登录的情况下，将显示登录页面，并展示当前在线用户的数量。
func Index(c *gin.Context) {
	c.Abort()
}

func Login(c *gin.Context) {
	user_service.Login(c)
}

func Logout(c *gin.Context) {
	user_service.Logout(c)
}

func Home(c *gin.Context) {
	userInfo := user_service.GetUserInfo(c)

	rooms := []map[string]interface{}{
		{"id": 1, "num": primary.OnlineRoomUserCount(1)},
		{"id": 2, "num": primary.OnlineRoomUserCount(2)},
		{"id": 3, "num": primary.OnlineRoomUserCount(3)},
		{"id": 4, "num": primary.OnlineRoomUserCount(4)},
		{"id": 5, "num": primary.OnlineRoomUserCount(5)},
		{"id": 6, "num": primary.OnlineRoomUserCount(6)},
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      rooms,
		"user_info": userInfo,
	})
}

// Room 函数用于显示指定房间的聊天室页面
// roomId 为由 URL 传入的房间号参数
func Room(c *gin.Context) {
	roomId := c.Param("room_id") // c.Param 用于获取 RESTful 风格的路径参数，例如 http://example.com/user/123中的 123。

	rooms := []string{"1", "2", "3", "4", "5", "6"}

	if !helper.InArray(roomId, rooms) {
		roomId = "1"
	}
	// 获取当前登录用户身份验证信息
	userInfo := user_service.GetUserInfo(c)
	// 获取指定房间中的历史聊天消息
	msgList := message_service.GetLimitMsg(roomId, 0)

	c.JSON(http.StatusOK, gin.H{
		"user_info":      userInfo,
		"msg_list":       msgList,
		"msg_list_count": len(msgList),
		"room_id":        roomId,
	})
}

// PrivateChat 函数用于显示两个用户之间的私聊页面
func PrivateChat(c *gin.Context) {

	// 从请求参数中读取聊天室 roomId 和聊天对象 toUid
	roomId := c.Query("room_id") // c.Query 用于获取 GET 请求中的 URL 参数，例如 http://example.com/?key=value 中的 key 和 value。
	toUid := c.Query("uid")

	userInfo := user_service.GetUserInfo(c)

	uid := strconv.Itoa(int(userInfo["uid"].(uint)))

	msgList := message_service.GetLimitPrivateMsg(uid, toUid, 0)

	c.HTML(http.StatusOK, "private_chat.html", gin.H{
		"user_info": userInfo,
		"msg_list":  msgList,
		"room_id":   roomId,
	})
}

// Pagination 函数用于获取分页数据，返回 JSON 格式的数据。
// 如果请求参数中的 room_id 不在允许列表中，则返回空数组。
func Pagination(c *gin.Context) {
	roomId := c.Query("room_id")
	toUid := c.Query("uid")
	offset := c.Query("offset")
	offsetInt, e := strconv.Atoi(offset)
	if e != nil || offsetInt <= 0 {
		offsetInt = 0
	}

	rooms := []string{"1", "2", "3", "4", "5", "6"}

	if !helper.InArray(roomId, rooms) {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": map[string]interface{}{
				"list": nil,
			},
		})
		return
	}

	var msgList []map[string]interface{}
	if toUid != "" {
		userInfo := user_service.GetUserInfo(c)

		uid := strconv.Itoa(int(userInfo["uid"].(uint)))

		msgList = message_service.GetLimitPrivateMsg(uid, toUid, offsetInt)
	} else {
		msgList = message_service.GetLimitMsg(roomId, offsetInt)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": map[string]interface{}{
			"list": msgList,
		},
	})
}
