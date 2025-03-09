package model

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `json:"id"`
	Age      int       `json:"age"`
	Login    string    `json:"login"`
	Password string    `json:"password"`
	Name     string    `json:"name"`
}

type UpdateUserRequest struct {
	ID       uuid.UUID `json:"id"`
	Age      int       `json:"age"`
	Password string    `json:"password"`
	Name     string    `json:"name"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
