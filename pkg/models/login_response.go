package models

type LoginResponse struct {
	SessionToken string `json:"session_token"`
	HtmlResult   string `json:"html_result"`
}
