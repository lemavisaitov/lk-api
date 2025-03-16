package model

import (
	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `json:"id" validate:"required,uuid"`
	Age      int       `json:"age" validate:"gte=0"`
	Login    string    `json:"login" validate:"required"`
	Password string    `json:"password" validate:"required"`
	Name     string    `json:"name" validate:"required"`
}

type UpdateUserRequest struct {
	ID       uuid.UUID `json:"id" validate:"uuid"`
	Age      int       `json:"age" validate:"gte=0"`
	Password string    `json:"password"`
	Name     string    `json:"name"`
}

type LoginRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}
