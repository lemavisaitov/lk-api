package model

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `json:"id"`
	Login    string    `json:"login"`
	Password string    `json:"email"`
	Name     string    `json:"name"`
	Age      int       `json:"age"`
}

type UpdateUserRequest struct {
	Password string `json:"password"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
}
