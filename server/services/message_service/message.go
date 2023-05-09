package message_service

import "via-chat-distributed/models"

func GetLimitMsg(roomId string, offset int) []models.MessageWithUserInfo {
	return models.GetLimitMsg(roomId, offset)
}

func GetLimitPrivateMsg(uid, toUId string, offset int) []models.MessageWithUserInfo {
	return models.GetLimitPrivateMsg(uid, toUId, offset)
}
