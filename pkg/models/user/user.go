package user

type User struct {
	Name         string `json:"name"`
	Username     string `json:"username"`
	SessionToken string `json:"session_token"`
}
