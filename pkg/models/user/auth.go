package user

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type User struct {
	Name      string `json:"name"`
	Username  string `json:"username"`
	SessionID string `json:"session_id"`
}

type UserInfo struct {
	Name     string `json:"name"`
	Username string `json:"username"`
}
