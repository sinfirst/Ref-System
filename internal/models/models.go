package models

type User struct {
	Username string `json:"login"`
	Password string `json:"password,omitempty"`
}
