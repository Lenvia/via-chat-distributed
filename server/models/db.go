package models

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

var ChatDB *gorm.DB
var RedisClient *redis.Client

const maxRetries = 5
const baseInterval = 5 * time.Second

func InitDB() {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	dsn := viper.GetString("mysql.dsn")
	var err error
	var retryInterval time.Duration
	for i := 0; i < maxRetries; i++ {
		ChatDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			AutoCreateTable()
			return
		}
		retryInterval = baseInterval * time.Duration(i+1)
		fmt.Printf("Failed to connect to MySQL, retrying in %v... (%v/%v)\n", retryInterval, i+1, maxRetries)
		time.Sleep(retryInterval) // 倍数回退
	}

	panic(fmt.Sprintf("Failed to connect to MySQL after %v retries: %v", maxRetries, err))
}

func AutoCreateTable() {
	err := ChatDB.AutoMigrate(&Message{})
	if err != nil {
		log.Println(err)
	}
	err = ChatDB.AutoMigrate(&User{})
	if err != nil {
		log.Println(err)
	}
}

func InitRedis() {
	addr := viper.GetString("redis.ip") + ":" + viper.GetString("redis.port")
	db := viper.GetInt("redis.db")
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // Redis 密码，如果没有设置可以为空字符串
		DB:       db, // Redis 数据库，默认为 0
	})

	_, err := RedisClient.Ping().Result()
	if err != nil {
		log.Println(err)
	} else {
		log.Println("redis connected.")
	}
}
