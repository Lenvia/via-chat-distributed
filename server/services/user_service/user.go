package user_service

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
	"via-chat-distributed/middleware"
	"via-chat-distributed/models"
	"via-chat-distributed/services/errmsg"
	"via-chat-distributed/services/helper"
	"via-chat-distributed/services/validator"
)

func Login(c *gin.Context) {
	var u validator.User

	if err := c.ShouldBind(&u); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 5000, "msg": err.Error()})
		return
	}

	username := u.Username
	pwd := u.Password
	avatarId := u.AvatarId
	encryptedPwd := helper.BcryptPwd(pwd) // pwd 是当前输入的密码

	user := models.FindUserByField("username", username)
	userInfo := user

	if userInfo.ID > 0 {
		// json 用户存在，验证密码
		// 注意，应该是输入的明文密码和 数据库里的hash字符串 进行验证
		PasswordErr := bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(pwd))

		if PasswordErr != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 5000,
				"msg":  "密码错误",
			})
			return
		}

		models.SaveAvatarId(avatarId, user)

	} else {
		// 新用户
		userInfo = models.AddUser(models.User{
			Username: username,
			Password: encryptedPwd,
			AvatarId: avatarId,
		})
	}

	if userInfo.ID > 0 {
		token, _ := getToken(userInfo) // 登录通过，返回token

		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"token": token,
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": 5001,
			"msg":  "系统错误",
		})
		return
	}
}

func GetUserInfo(c *gin.Context) map[string]interface{} {
	username, _ := c.Value("username").(string)

	data := make(map[string]interface{})

	if username != "" { // 使用此ID检索用户信息，例如：ID，用户名和头像编号
		user := models.FindUserByField("username", username)
		data["uid"] = user.ID // 这里还没做转换是因为 data["uid"] 字段与 user.ID column 对应 "id" 字段不一致
		data["username"] = user.Username
		data["avatar_id"] = user.AvatarId
	}

	return data
}

func Logout(c *gin.Context) {
	c.Set("username", "")
	c.Redirect(http.StatusFound, "/")
	return
}

func getToken(user models.User) (token string, errMsg string) {
	j := middleware.NewJWT()
	claims := middleware.MyClaims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix() - 100,
			ExpiresAt: time.Now().Unix() + 604800,
			Issuer:    "viaChat",
		},
	}

	token, err := j.CreateToken(claims)
	if err != nil {
		return "", errmsg.GetErrMsg(errmsg.ERROR)
	}

	return token, errmsg.GetErrMsg(errmsg.SUCCESS)

}
