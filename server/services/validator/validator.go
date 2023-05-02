package validator

type User struct {
	Username string `binding:"required,max=16,min=2" form:"username" json:"username"`
	Password string `binding:"required,max=32,min=6" form:"password" json:"password"`
	AvatarId string `binding:"required,numeric" form:"avatar_id" json:"avatar_id"`
}
