package models

type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
	RoomKey string `json:"room_key"`
	HeadUrl string `json:"head_url"`
}
