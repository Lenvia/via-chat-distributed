package models

import (
	"gorm.io/gorm"
	"log"
	"time"
)

type User struct {
	gorm.Model
	ID        uint      `gorm:"column:id"`
	Username  string    `gorm:"column:username"`
	Password  string    `gorm:"column:password"`
	AvatarId  string    `gorm:"column:avatar_id"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func AddUser(value interface{}) User {
	var u User
	u.Username = value.(map[string]interface{})["username"].(string)
	u.Password = value.(map[string]interface{})["password"].(string)
	u.AvatarId = value.(map[string]interface{})["avatar_id"].(string)

	tx := ChatDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 发生错误时回滚事务
		}
	}()

	err := tx.Create(&u).Error
	if err != nil {
		tx.Rollback() // 发生错误时回滚事务
		log.Println(err)
	}

	// 提交事务
	err = tx.Commit().Error
	if err != nil {
		log.Println(err)
	}
	return u
}

func SaveAvatarId(AvatarId string, u User) User {
	u.AvatarId = AvatarId

	tx := ChatDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 发生错误时回滚事务
		}
	}()

	err := tx.Save(&u).Error
	if err != nil {
		tx.Rollback() // 发生错误时回滚事务
		log.Println(err)
	}

	// 提交事务
	err = tx.Commit().Error
	if err != nil {
		log.Println(err)
	}

	return u
}

func FindUserByField(field, value string) User {
	var u User

	if field == "id" || field == "username" {
		ChatDB.Where(field+" = ?", value).First(&u)
	}

	return u
}

func GetOnlineUserList(uids []float64) []map[string]interface{} {
	var results []map[string]interface{}
	ChatDB.Where("id IN ?", uids).Find(&results)

	return results
}
