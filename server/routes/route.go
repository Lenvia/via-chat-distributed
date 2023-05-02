package routes

import (
	"github.com/gin-gonic/gin"
	"via-chat/api/v1"
	"via-chat/middleware"
	"via-chat/ws/primary"
)

func InitRoute() *gin.Engine {
	//router := gin.Default()
	router := gin.New()
	router.Use(middleware.Cors())

	//if viper.GetString(`app.debug_mod`) == "false" {
	//	// live 模式 打包用，使用嵌入式的静态资源FS
	//	router.StaticFS("/static", http.FS(static.EmbedStatic))
	//} else {
	//	// dev 开发用 避免修改静态资源需要重启服务，使用本地的静态文件目录
	//	router.StaticFS("/static", http.Dir("static"))
	//}

	// 创建路由分组，并启用 cookie-based 会话
	sr := router.Group("/")
	{
		sr.GET("/", v1.Index)

		sr.POST("/login", v1.Login)
		sr.GET("/logout", v1.Logout)
		sr.GET("/ws", primary.Start)

		authorized := sr.Group("/")
		authorized.Use(middleware.JwtToken())
		{
			authorized.GET("/home", v1.Home)
			authorized.GET("/room/:room_id", v1.Room)
			authorized.GET("/private-web", v1.PrivateChat)
			//authorized.POST("/img-kr-upload", v1.ImgKrUpload)
			authorized.GET("/pagination", v1.Pagination)
		}

	}

	return router
}
