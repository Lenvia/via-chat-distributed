package helper

import (
	"golang.org/x/crypto/bcrypt"
	"log"
	"unicode/utf8"
)

func InArray(needle interface{}, hystack interface{}) bool {
	switch key := needle.(type) {
	case string:
		for _, item := range hystack.([]string) {
			if key == item {
				return true
			}
		}
	case int:
		for _, item := range hystack.([]int) {
			if key == item {
				return true
			}
		}
	case int64:
		for _, item := range hystack.([]int64) {
			if key == item {
				return true
			}
		}
	default:
		return false
	}
	return false
}

func BcryptPwd(pwd string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalln(err)
	}
	return string(hashedPassword)
}

func MbStrLen(str string) int {
	return utf8.RuneCountInString(str)
}
