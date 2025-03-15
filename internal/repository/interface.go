package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/lemavisaitov/lk-api/internal/model"
)

type UserProvider interface {
	UpdateUser(context.Context, model.UpdateUserRequest) (*uuid.UUID, error)
	AddUser(context.Context, model.User) error
	GetUser(context.Context, uuid.UUID) (*model.User, error)
	GetUserIDByLogin(context.Context, string) (*uuid.UUID, error)
	DeleteUser(context.Context, uuid.UUID) error
}
