package user

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type User struct {
	Name      string `json:"name" bson:"name"`
	Username  string `json:"username" bson:"username"`
	SessionID string `json:"session_id" bson:"session_id"`
}
