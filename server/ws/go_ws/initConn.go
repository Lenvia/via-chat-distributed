package go_ws

import (
	"github.com/robfig/cron/v3"
)

func InitConn() {
	CleanOfflineConn()

	go Write() // 必须创建一个额外的协程来不断从 SMsg 中取消息，使服务器连接为空时也不会阻塞
}

func CleanOfflineConn() {

	c := cron.New()

	// 每天定时执行的条件
	spec := `* * * * *`

	c.AddFunc(spec, func() {
		// fmt.Println("CleanOfflineConn")
		HandelOfflineCoon()
	})

	go c.Start()
}
