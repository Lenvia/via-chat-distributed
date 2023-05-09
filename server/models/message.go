package models

import (
	"encoding/json"
	"gorm.io/gorm"
	"log"
	"sort"
	"time"
)

type Message struct { // 这里的json tag 是用于 redis 序列化
	gorm.Model
	ID        uint      `gorm:"column:id" json:"id"`
	UserId    int       `gorm:"column:user_id" json:"user_id"`
	ToUserId  int       `gorm:"column:to_user_id" json:"to_user_id"`
	RoomId    int       `gorm:"column:room_id" json:"room_id"`
	Content   string    `gorm:"column:content" json:"content"`
	ImageUrl  string    `gorm:"column:image_url" json:"image_url"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type MessageWithUserInfo struct {
	Message
	Username string `gorm:"column:username" json:"username"`
	AvatarID string `gorm:"column:avatar_id" json:"avatar_id"`
}

// SaveContent 函数将消息内容保存到数据库中，value 参数为消息内容的 map 类型数据。
func SaveContent(m Message) Message {
	tx := ChatDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 发生错误时回滚事务
		}
	}()

	err := tx.Create(&m).Error
	if err != nil {
		tx.Rollback() // 发生错误时回滚事务
		log.Println(err)
	}

	// 提交事务
	err = tx.Commit().Error
	if err != nil {
		log.Println(err)
	}

	return m
}

// GetLimitMsg 函数从数据库中查询指定房间的聊天记录，roomId 参数为房间 ID，offset 参数为分页偏移量，返回值为查询结果的 map 切片。
func GetLimitMsg(roomId string, offset int) []MessageWithUserInfo {
	var results []MessageWithUserInfo

	// 首先从redis读取消息
	key := "room:" + roomId + ":messages"
	messages, err := RedisClient.LRange(key, 0, -1).Result()
	if err != nil || len(messages) == 0 {
		// 如果 Redis List 中不存在历史消息，则从数据库中查询历史消息
		// 注意这里第一个是 Message！因为没有 MessageWithUserInfo 这个表
		ChatDB.Model(&Message{}).
			Select("messages.*, users.username ,users.avatar_id").
			Joins("INNER Join users on users.id = messages.user_id").
			Where("messages.room_id = " + roomId).
			Where("messages.to_user_id = 0").
			Order("messages.id desc").
			Offset(offset).
			Limit(100).
			Scan(&results)

		// 如果 offset 为 0，则按照 id 升序排序
		if offset == 0 {
			sort.Slice(results, func(i, j int) bool {
				return results[i].ID < results[j].ID
			})
		}

		// 将历史消息写入 Redis List 中
		for _, msgWithU := range results {
			jsonBytes, err := json.Marshal(msgWithU)
			if err != nil {
				log.Println(err)
			}
			jsonString := string(jsonBytes)

			err = RedisClient.RPush(key, jsonString).Err()
			if err != nil {
				log.Println(err)
			}
		}
	} else {
		for _, message := range messages {
			var msgWithU MessageWithUserInfo
			err := json.Unmarshal([]byte(message), &msgWithU)
			if err != nil {
				log.Println(err)
				continue
			}
			results = append(results, msgWithU)
		}
	}

	return results
}

func GetLimitPrivateMsg(uid, toUId string, offset int) []MessageWithUserInfo {

	var results []MessageWithUserInfo
	ChatDB.Model(&Message{}).
		Select("messages.*, users.username ,users.avatar_id").
		Joins("INNER Join users on users.id = messages.user_id").
		Where("(" +
			"(" + "messages.user_id = " + uid + " and messages.to_user_id=" + toUId + ")" +
			" or " +
			"(" + "messages.user_id = " + toUId + " and messages.to_user_id=" + uid + ")" +
			")").
		Order("messages.id desc").
		Offset(offset).
		Limit(100).
		Scan(&results)

	if offset == 0 {
		sort.Slice(results, func(i, j int) bool {
			return results[i].ID < results[j].ID
		})
	}

	return results
}
