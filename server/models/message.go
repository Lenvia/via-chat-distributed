package models

import (
	"gorm.io/gorm"
	"log"
	"sort"
	"strconv"
	"time"
)

// Message TODO 可以使用 mapstructure.Decode() 函数将 map 解码为 Go 结构体，从而省略在 SaveContent() 函数中使用断言的部分。
type Message struct {
	gorm.Model
	ID        uint      `gorm:"column:id"`
	UserId    int       `gorm:"column:user_id"`
	ToUserId  int       `gorm:"column:to_user_id"`
	RoomId    int       `gorm:"column:room_id"`
	Content   string    `gorm:"column:content"`
	ImageUrl  string    `gorm:"column:image_url"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// SaveContent 函数将消息内容保存到数据库中，value 参数为消息内容的 map 类型数据。
func SaveContent(value interface{}) Message {
	var m Message
	m.UserId = value.(map[string]interface{})["user_id"].(int)
	m.ToUserId = value.(map[string]interface{})["to_user_id"].(int)
	m.Content = value.(map[string]interface{})["content"].(string)

	roomIdStr := value.(map[string]interface{})["room_id"].(string)

	roomIdInt, _ := strconv.Atoi(roomIdStr)

	m.RoomId = roomIdInt

	if _, ok := value.(map[string]interface{})["image_url"]; ok {
		m.ImageUrl = value.(map[string]interface{})["image_url"].(string)
	}

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
func GetLimitMsg(roomId string, offset int) []map[string]interface{} {

	var results []map[string]interface{}
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
			return results[i]["id"].(uint64) < results[j]["id"].(uint64)
		})
	}

	return results
}

func GetLimitPrivateMsg(uid, toUId string, offset int) []map[string]interface{} {

	var results []map[string]interface{}
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
			return results[i]["id"].(uint64) < results[j]["id"].(uint64)
		})
	}

	return results
}
