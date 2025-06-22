package models

type User struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Role  int64  `json:"role"`
}
