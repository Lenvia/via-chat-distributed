package message_service

import "via-chat-distributed/models"

func GetLimitMsg(roomId string, offset int) []map[string]interface{} {
	return models.GetLimitMsg(roomId, offset)
}

func GetLimitPrivateMsg(uid, toUId string, offset int) []map[string]interface{} {
	return models.GetLimitPrivateMsg(uid, toUId, offset)
}
