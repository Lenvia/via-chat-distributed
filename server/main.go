package main

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"via-chat-distributed/configs"
	"via-chat-distributed/models"
	"via-chat-distributed/routes"
	"via-chat-distributed/ws/go_ws"
)

func init() {
	// 设置配置文件类型为 JSON
	viper.SetConfigType("json")

	// 读取配置文件，如果文件不可用则记录日志
	if err := viper.ReadConfig(bytes.NewBuffer(configs.AppJsonConfig)); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("no such config file") // 配置文件不存在，记录日志
		} else {
			log.Println("read config error") // 配置文件存在但读取出错，记录日志
		}
		log.Fatal(err) // 读取配置文件失败，记录日志并退出程序
	}

	models.InitDB()           // mysql
	models.InitRedis()        // redis
	models.InitGptClient()    // gpt
	models.InitMessageQueue() // 消息队列

}

func main() {
	gin.SetMode(gin.DebugMode) // 设置 Gin 框架为调试模式

	port := viper.GetString(`app.port`) // 获取应用程序端口号

	// 初始化路由
	// 当用户连接到应用程序时，调用 routes 中的 primary.Start 方法创建 WebSocket 连接，并将其绑定到 "/ws" 路径上。
	// 对于每个连接，都会为其创建一个独立的 WebSocket 连接对象，并保存到连接池中。不同的用户连接之间不会发生干扰
	router := routes.InitRoute()

	//router.SetHTMLTemplate(views.GoTpl) // 加载模板文件

	go_ws.InitConn() // 清理已离线的 WebSocket 连接

	log.Println("listening port:", port) // 打印监听端口的地址信息

	http.ListenAndServe(":"+port, router) // 启动 HTTP 服务器并监听端口
}
