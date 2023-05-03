package middleware

import (
	"errors"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"strings"
	"via-chat-distributed/services/errmsg"
)

type JWT struct {
	JWTKey []byte
}

func NewJWT() *JWT {
	return &JWT{[]byte(viper.GetString("jwt_key"))}
}

type MyClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var (
	TokenExpired     = errors.New("token expired")
	TokenNotValidYet = errors.New("token not valid")
	TokenMalformed   = errors.New("token incorrect")
	TokenInvalid     = errors.New("not a token")
)

// CreateToken 生成token
func (j *JWT) CreateToken(claims MyClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.JWTKey)
}

func (j *JWT) ParseToken(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.JWTKey, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 { // token 不正确
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 { // 已过期
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 { // 无效
				return nil, TokenNotValidYet
			} else { // 非token
				return nil, TokenInvalid
			}

		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*MyClaims); ok && token.Valid { // 解析成功
			return claims, nil
		}
	}
	return nil, TokenInvalid
}

// JwtToken jwt中间件
func JwtToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		tokenHeader := c.Request.Header.Get("Authorization")

		if tokenHeader == "" { // 没有token
			code = errmsg.ErrorTokenExist
			c.JSON(http.StatusOK, gin.H{
				"status":  code,
				"message": errmsg.GetErrMsg(code),
			})
			c.Abort()
			return
		}

		checkToken := strings.Split(tokenHeader, " ")
		if len(checkToken) == 0 {
			code = errmsg.ErrorTokenWrong
			c.JSON(http.StatusOK, gin.H{
				"status":  code,
				"message": errmsg.GetErrMsg(code),
			})
			c.Abort()
			return
		}

		if len(checkToken) != 2 || checkToken[0] != "Bearer" {
			code = errmsg.ErrorTokenTypeWrong
			c.JSON(http.StatusOK, gin.H{
				"status":  code,
				"message": errmsg.GetErrMsg(code),
			})
			c.Abort()
			return
		}

		// Token 格式正确，验证 token
		j := NewJWT()
		claims, err := j.ParseToken(checkToken[1])
		if err != nil {
			if err == TokenExpired {
				c.JSON(http.StatusOK, gin.H{
					"status":  errmsg.ERROR,
					"message": TokenExpired.Error(),
					"data":    nil,
				})
				c.Abort()
				return
			}
			// 其他错误
			c.JSON(http.StatusOK, gin.H{
				"status":  errmsg.ERROR,
				"message": err.Error(),
				"data":    nil,
			})
			c.Abort()
			return
		}
		// 通过验证
		c.Set("username", claims.Username)
		c.Next()
	}
}
