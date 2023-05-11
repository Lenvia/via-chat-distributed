package routes

import (
	"github.com/gin-gonic/gin"
	"via-chat-distributed/api/v1"
	"via-chat-distributed/middleware"
	"via-chat-distributed/ws/go_ws"
)

func InitRoute() *gin.Engine {
	//router := gin.Default()
	router := gin.New()
	router.Use(middleware.Cors())

	// 创建路由分组，并启用 cookie-based 会话
	sr := router.Group("/")
	{
		sr.GET("/", v1.Index)

		sr.POST("/login", v1.Login)
		sr.GET("/logout", v1.Logout)
		sr.GET("/ws", go_ws.Start)

		authorized := sr.Group("/")
		authorized.Use(middleware.JwtToken())
		{
			authorized.GET("/home", v1.Home)
			authorized.GET("/room/:room_id", v1.Room)
			authorized.GET("/private-web", v1.PrivateChat)
			//authorized.POST("/img-kr-upload", v1.ImgKrUpload)
			//authorized.GET("/pagination", v1.Pagination)
		}

	}

	return router
}
