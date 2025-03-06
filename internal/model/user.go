package model

import "github.com/google/uuid"

type User struct {
	ID       uuid.NullUUID `json:"id"`
	Login    string        `json:"login"`
	Password string        `json:"email"`
	Name     string        `json:"name"`
	Age      int           `json:"age"`
}

type UpdateUserRequest struct {
	ID       uuid.UUID `json:"id"`
	Password string    `json:"password"`
	Name     string    `json:"name"`
	Age      int       `json:"age"`
}
