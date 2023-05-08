package go_ws

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jianfengye/collection"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"via-chat-distributed/models"
	"via-chat-distributed/pb/gpt"
	"via-chat-distributed/services/helper"
	"via-chat-distributed/ws"
)

// 客户端连接详情

// WsClients 结构体定义了 WebSocket 客户端的信息
type WsClients struct {
	Conn       *websocket.Conn // websocket 连接对象
	RemoteAddr string          // 客户端远程地址
	Uid        string          // 客户端唯一标识符
	Username   string          // 客户端用户名
	RoomId     string          // 客户端所在房间 id
	AvatarId   string          // 客户端头像 id
}

// pingStorage 结构体定义了心跳包信息
type pingStorage struct {
	Conn       *websocket.Conn // websocket 连接对象
	RemoteAddr string          // 客户端远程地址
	Time       int64           // 心跳包发送时间
}

// 全局变量定义初始化（当前server服务的所有用户共享）
var (
	wsUpgrader = websocket.Upgrader{} // WebSocket 升级器，用于升级普通的 HTTP 连接为 WebSocket 连接
	//clientMsg  = models.WebSocketMsg{}
	mutex = sync.Mutex{}

	// rooms = [roomCount + 1][]WsClients{}
	rooms       = make(map[int][]interface{}) // 聊天室 map，以房间 id 为 key，保存连接对象和其他客户端信息
	conn2roomId = make(map[*websocket.Conn]string)
	enterRooms  = make(chan WsClients)       // 进入聊天室的客户端连接，用于处理客户端连接请求
	offline     = make(chan *websocket.Conn) // 离线客户端的连接，用于处理客户端断开连接的请求
	chNotify    = make(chan int, 1)          // 通知客户端，用于处理对聊天室客户端状态变化的通知
	pingMap     []interface{}                // 心跳列表，存储客户端的心跳检测信息
)

// 定义消息类型
const msgTypeOnline = 1        // 上线
const msgTypeOffline = 2       // 离线
const msgTypeSend = 3          // 消息发送
const msgTypeGetOnlineUser = 4 // 获取用户列表
const msgTypePrivateChat = 5   // 私聊

const roomCount = 6 // 房间总数

func publishMsg(serializedMsg []byte) {
	err := models.NC.Publish(models.BaseTopic, serializedMsg)
	if err != nil {
		log.Println(err)
	}
}

// ---------------------------------------------------------------------------------

type GoServe struct {
	ws.ServeInterface
}

func (goServe *GoServe) RunWs(gin *gin.Context) {
	// 使用 channel goroutine
	Run(gin)
}

func (goServe *GoServe) GetOnlineUserCount() int {
	return GetOnlineUserCount()
}

func (goServe *GoServe) GetOnlineRoomUserCount(roomId int) int {
	return GetOnlineRoomUserCount(roomId)
}

func Run(gin *gin.Context) {

	// @see https://github.com/gorilla/websocket/issues/523
	// wsUpgrader.CheckOrigin 是用来解决 websocket 跨域问题的，这里设置为返回 true，表示接收来自任何源的请求。
	wsUpgrader.CheckOrigin = func(r *http.Request) bool { return true }

	c, _ := wsUpgrader.Upgrade(gin.Writer, gin.Request, nil)

	defer c.Close()

	go read(c)

	// 对于每一个客户端连接，也会新建一个协程去监听 enterRooms 和 models.SMsg 这两个通道。
	// 多个协程可以并发读写通道，但在任意时刻，只有其中的一个协程可以读取或写入该通道
	go Write()

	select {} // 在无限循环中等待客户端的响应，这是阻塞的。当读协程或写协程的通道收到信息时，将继续进行操作。

}

// HandelOfflineCoon 定时任务清理没有心跳的连接
func HandelOfflineCoon() {
	// 通过 collection 包的 NewObjCollection 函数，将 pingMap 转化为一个可操作的 collection 对象
	objColl := collection.NewObjCollection(pingMap)

	// 使用 Reject 方法遍历 pingMap，过滤出时间差超过 60 秒的不活跃客户端，并将其断开连接
	// retColl 保存 objColl.Reject(返回值为false) 的连接集合，即活跃的
	retColl := objColl.Reject(func(obj interface{}, index int) bool {
		nowTime := time.Now().Unix()
		timeDiff := nowTime - obj.(pingStorage).Time
		log.Println("timeDiff", nowTime, obj.(pingStorage).Time, timeDiff)

		if timeDiff > 60 { // 超过 60s 没有心跳 主动断开连接
			offline <- obj.(pingStorage).Conn // 将该客户端的连接对象添加到 offline 通道中，等待下一次检查时断开连接
			return true
		}
		return false
	})

	// 将处理后的 collection 对象转化为其他的 interface{} 类型的 slice，用于更新 pingMap
	interfaces, _ := retColl.ToInterfaces()

	// 更新 pingMap，删除不活跃的客户端
	pingMap = interfaces
}

// appendPing 函数用于在 pingMap 中添加新的客户端信息，实现心跳机制
func appendPing(c *websocket.Conn) {
	objColl := collection.NewObjCollection(pingMap)

	// 删除已经存在的与新连接相同的客户端信息
	retColl := objColl.Reject(func(obj interface{}, index int) bool {
		if obj.(pingStorage).RemoteAddr == c.RemoteAddr().String() {
			return true
		}
		return false
	})

	// 再追加
	retColl.Append(pingStorage{
		Conn:       c,
		RemoteAddr: c.RemoteAddr().String(),
		Time:       time.Now().Unix(),
	})

	interfaces, _ := retColl.ToInterfaces()
	pingMap = interfaces

}

func read(c *websocket.Conn) {
	defer func() {
		//捕获read抛出的panic
		if err := recover(); err != nil {
			log.Println("read发生错误", err)
			//panic(nil)
		}
	}()

	for { // 循环，不断读取客户端发来的消息
		_, message, err := c.ReadMessage()
		if err != nil { // 离线通知
			// 将该客户端的连接对象添加到 offline 通道中，等待下一次检查时断开连接
			offline <- c
			log.Println("ReadMessage error1", err)
			return
		}

		serveMsgStr := message

		// 处理心跳响应 , heartbeat为与客户端约定的值
		if string(serveMsgStr) == `heartbeat` {
			appendPing(c)
			log.Println(pingMap)
			chNotify <- 1
			c.WriteMessage(websocket.TextMessage, []byte(`{"status":0,"data":"heartbeat ok"}`)) // 向客户端发送心跳响应
			<-chNotify
			continue
		}

		// 最关键的地方！这里谨慎变更
		var clientMsg models.WebSocketMsg
		json.Unmarshal(message, &clientMsg)
		fmt.Println("来自客户端的消息", clientMsg, c.RemoteAddr())
		if clientMsg.Data.Uid != "" { // 已经登录过的用户
			if clientMsg.Status == msgTypeOnline { // 进入房间，建立连接
				enterRooms <- WsClients{
					Conn:       c,
					RemoteAddr: c.RemoteAddr().String(),
					Uid:        clientMsg.Data.Uid,
					Username:   clientMsg.Data.Username,
					RoomId:     clientMsg.Data.RoomId,
					AvatarId:   clientMsg.Data.AvatarId,
				}
			}

			// 根据客户端发送的消息类型，将其转化为需要发送给其他客户端的服务端消息，并添加到消息队列中，等待发送
			serveMsgBytes, _ := formatServeMsgStr(&clientMsg)
			publishMsg(serveMsgBytes)

			//models.SMsg <- serveMsg
			go requestGPT(&clientMsg)
		}
	}
}

// Write 函数是单独在一个 goroutine 中执行的，用于向所有 WebSocket 客户端发送消息
func Write() {
	defer func() {
		//捕获write抛出的panic
		if err := recover(); err != nil {
			log.Println("write发生错误", err)

		}
	}()

	for {
		select {
		// 如果从 enterRooms 通道中获取到一个客户端连接信息，则处理该连接
		case r := <-enterRooms:
			handleConnClients(&r)
		// 如果从 models.SMsg 通道中获取到一个服务端消息，则将其转化为需要发送给客户端的 JSON 字符串，并根据不同的消息类型进行相应的处理
		case cl := <-models.SMsg:
			fmt.Println("即将发送消息：", cl)
			serveMsgStr, _ := json.Marshal(cl)
			switch cl.Status {
			// 如果是在线消息或者发送消息，则向所有的客户端发送该消息
			case msgTypeOnline, msgTypeSend:
				notify(string(serveMsgStr), cl.Data.RoomId) // 发送者，发送消息

				//case msgTypeGetOnlineUser:
				//	// 无缓冲区通道 chNotify 确保同一时刻只有一个协程向客户端发送消息
				//	chNotify <- 1
				//	cl.Conn.WriteMessage(websocket.TextMessage, serveMsgStr)
				//	<-chNotify
				//case msgTypePrivateChat:
				//	chNotify <- 1
				//	toC := findToUserCoonClient() // 查找需要发送消息的客户端连接对象，并发送消息
				//	if toC != nil {
				//		toC.(WsClients).Conn.WriteMessage(websocket.TextMessage, serveMsgStr)
				//	}
				//	<-chNotify
			}
		case o := <-offline:
			disconnect(o)
		}
	}
}

func handleConnClients(clients *WsClients) {
	c := clients.Conn
	roomId := clients.RoomId

	roomIdInt, _ := strconv.Atoi(roomId)
	conn2roomId[c] = roomId

	objColl := collection.NewObjCollection(rooms[roomIdInt])

	// 使用 objColl.Reject 过滤出不是当前客户端的连接对象
	// 最终结果返回的是一个不包含已有同样 UID 连接的连接集合。
	retColl := objColl.Reject(func(item interface{}, key int) bool {
		if item.(WsClients).Uid == clients.Uid {
			// 如果已有同样的UID连接，则向该连接发送无效的错误消息，并返回 true
			item.(WsClients).Conn.WriteMessage(websocket.TextMessage, []byte(`{"status":-1,"data":[]}`))
			return true
		}
		return false
	})

	// 将当前用户信息添加到 retColl 中
	retColl.Append(WsClients{
		Conn:       c,
		RemoteAddr: c.RemoteAddr().String(),
		Uid:        clients.Uid,
		Username:   clients.Username,
		RoomId:     roomId,
		AvatarId:   clients.AvatarId,
	})

	interfaces, _ := retColl.ToInterfaces()

	// 更新 rooms 对应房间中存储的连接对象集合
	rooms[roomIdInt] = interfaces

	//mutex.Lock()

	//mutex.Unlock()
}

// findToUserCoonClient 获取私聊的用户连接
//func findToUserCoonClient() interface{} {
//	_, roomIdInt := getRoomId(clientMsg)
//
//	toUserUid := clientMsg.Data.ToUid
//	assignRoom := rooms[roomIdInt]
//	for _, c := range assignRoom {
//		stringUid := c.(WsClients).Uid
//		if stringUid == toUserUid {
//			return c
//		}
//	}
//
//	return nil
//}

// notify 函数用于向所有连接到同一个房间的客户端发送消息
func notify(msg string, roomId string) {
	chNotify <- 1 // 利用channel阻塞 避免并发去对同一个连接发送消息出现panic: concurrent write to websocket connection这样的异常
	roomIdInt, _ := strconv.Atoi(roomId)
	assignRoom := rooms[roomIdInt]
	//fmt.Println("将要广播的房间号为：", roomIdInt)
	// 遍历该房间中所有的客户端连接对象，并向除了当前连接对象之外的其它客户端连接对象发送消息
	//fmt.Println("当前房间的连接：", assignRoom)
	for _, client := range assignRoom {
		fmt.Println(client.(WsClients).RemoteAddr)
		client.(WsClients).Conn.WriteMessage(websocket.TextMessage, []byte(msg))
	}
	fmt.Println("发送成功")
	fmt.Println()
	<-chNotify
}

// 离线通知
func disconnect(conn *websocket.Conn) {
	roomId := conn2roomId[conn]
	roomIdInt, _ := strconv.Atoi(roomId)

	// 创建一个通用对象集合，存储当前房间的所有连接对象
	objColl := collection.NewObjCollection(rooms[roomIdInt])

	// 过滤出需离开的连接对象
	retColl := objColl.Reject(func(item interface{}, key int) bool {
		// 如果当前连接的RemoteAddr和item的RemoteAddr相同，则执行对应的离线流程
		if item.(WsClients).RemoteAddr == conn.RemoteAddr().String() {
			data := models.MsgData{
				Username: item.(WsClients).Username,
				Uid:      item.(WsClients).Uid,
				Time:     time.Now().UnixNano() / 1e6, // 13位  10位 => now.Unix()
				RoomId:   roomId,
			}

			jsonStrServeMsg := models.WebSocketMsg{
				Status: msgTypeOffline,
				Data:   data,
			}
			serveMsgStr, _ := json.Marshal(jsonStrServeMsg)
			disMsg := string(serveMsgStr)

			// 关闭连接，并向整个房间的在线连接发送离线通知消息
			item.(WsClients).Conn.Close()
			notify(disMsg, roomId)
			return true
		}
		return false
	})

	// 将过滤后的连接对象重新转换为接口类型的切片，并更新 rooms 对应房间中存储的连接对象集合
	interfaces, _ := retColl.ToInterfaces()
	rooms[roomIdInt] = interfaces
	delete(conn2roomId, conn)
}

// 格式化传送给客户端的消息数据
func formatServeMsgStr(clientMsg *models.WebSocketMsg) ([]byte, models.WebSocketMsg) {
	roomId, roomIdInt := getRoomId(clientMsg)
	status := clientMsg.Status

	//log.Println(reflect.TypeOf(var))

	data := models.MsgData{ // 需要制作的消息
		Username: clientMsg.Data.Username,
		Uid:      clientMsg.Data.Uid,
		RoomId:   roomId,
		Time:     time.Now().UnixNano() / 1e6, // 13位  10位 => now.Unix()
	}

	// 如果消息类型是发送消息或私聊消息
	if status == msgTypeSend || status == msgTypePrivateChat {
		data.AvatarId = clientMsg.Data.AvatarId
		content := clientMsg.Data.Content

		data.Content = content
		if helper.MbStrLen(content) > 800 {
			// 如果内容的长度超过800个字符，则将其截断
			data.Content = string([]rune(content)[:800])
		}

		data.ToUid = clientMsg.Data.ToUid
		toUid, _ := strconv.Atoi(data.ToUid)
		intUid, _ := strconv.Atoi(data.Uid)

		msg := models.Message{
			UserId:   intUid,
			ToUserId: toUid,
			Content:  content,
			RoomId:   roomIdInt,
		}
		if clientMsg.Data.ImageUrl != "" {
			// 存在图片，同时保存消息的图片信息
			msg.ImageUrl = clientMsg.Data.ImageUrl
		}

		msg = models.SaveContent(msg)

		// 创建时间封装进去，发送回客户端
		data.CreatedAt = msg.CreatedAt
		data.UpdatedAt = msg.UpdatedAt
		data.ID = msg.ID

	}
	// 如果消息类型是获取在线用户列表
	if status == msgTypeGetOnlineUser {
		ro := rooms[roomIdInt]
		data.Count = len(ro)
		data.List = ro
	}

	jsonStrServeMsg := models.WebSocketMsg{
		Status: status,
		Data:   data,
	}
	serveMsgStr, _ := json.Marshal(jsonStrServeMsg)

	return serveMsgStr, jsonStrServeMsg
}

func getRoomId(clientMsg *models.WebSocketMsg) (string, int) {
	roomId := clientMsg.Data.RoomId

	roomIdInt, _ := strconv.Atoi(roomId)
	return roomId, roomIdInt
}

func requestGPT(clientMsg *models.WebSocketMsg) {
	fmt.Println(clientMsg.Data.Content)
	pattern := "@GPT"
	var reply *gpt.GptMsgResponse
	var err error
	if strings.HasPrefix(clientMsg.Data.Content, pattern) {
		query := clientMsg.Data.Content[len(pattern):]
		if models.GptClient != nil {
			reply, err = models.GptClient.Send(context.Background(), &gpt.GptMsgRequest{Query: query})
			if err != nil {
				log.Println(err)
				return
			}

			roomId, roomIdInt := getRoomId(clientMsg)
			// 持久化
			message := models.Message{
				UserId:   models.ChatGptIdInt,
				ToUserId: 0,
				Content:  reply.String(),
				RoomId:   roomIdInt,
			}

			message = models.SaveContent(message)

			// 制作消息
			data := models.MsgData{
				Username:  models.ChatGptName,
				Uid:       strconv.Itoa(models.ChatGptIdInt),
				RoomId:    roomId,
				Content:   reply.String(),
				Time:      time.Now().UnixNano() / 1e6, // 13位  10位 => now.Unix()
				CreatedAt: message.CreatedAt,
				UpdatedAt: message.UpdatedAt,
				ID:        message.ID,
			}

			jsonStrServeMsg := models.WebSocketMsg{
				Status: msgTypeSend,
				Data:   data,
			}
			//models.SMsg <- jsonStrServeMsg
			serveMsgBytes, _ := json.Marshal(jsonStrServeMsg)
			publishMsg(serveMsgBytes)
		}
	}

}

// =======================对外方法=====================================

func GetOnlineUserCount() int {
	num := 0
	for i := 1; i <= roomCount; i++ {
		num = num + GetOnlineRoomUserCount(i)
	}
	return num
}

func GetOnlineRoomUserCount(roomId int) int {
	return len(rooms[roomId])
}
