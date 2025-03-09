package usecase

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lemavisaitov/lk-api/internal/model"
	"github.com/lemavisaitov/lk-api/internal/repository"
	"github.com/pkg/errors"
)

type UserProvider interface {
	AddUser(*gin.Context, model.User) (uuid.UUID, error)
	GetUser(*gin.Context, uuid.UUID) (model.User, error)
	GetUserIDByLogin(*gin.Context, string) (uuid.UUID, error)
	UpdateUser(*gin.Context, model.UpdateUserRequest) (uuid.UUID, error)
	DeleteUser(*gin.Context, uuid.UUID) error
	LoginExists(*gin.Context, string) (bool, error)
}

type UserCase struct {
	userRepo repository.UserProvider
}

func NewUserCase(userRepo repository.UserProvider) *UserCase {
	return &UserCase{userRepo: userRepo}
}

func (u *UserCase) AddUser(c *gin.Context, user model.User) (uuid.UUID, error) {
	id, err := uuid.NewV7()

	if err != nil {
		return id, errors.Wrap(err, "error creating uuid")
	}

	user.ID.UUID = id
	if err := u.userRepo.AddUser(c, user); err != nil {
		return id, err
	}
	return id, nil
}

func (u *UserCase) GetUser(c *gin.Context, userID uuid.UUID) (model.User, error) {
	user, err := u.userRepo.GetUser(c, userID)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (u *UserCase) UpdateUser(c *gin.Context, req model.UpdateUserRequest) (uuid.UUID, error) {
	id, err := u.userRepo.UpdateUser(c, req)
	if err != nil {
		return id, errors.Wrap(err, "error updating user")
	}
	return id, nil
}

func (u *UserCase) GetUserIDByLogin(c *gin.Context, login string) (uuid.UUID, error) {
	id, err := u.userRepo.GetUserIDByLogin(c, login)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (u *UserCase) DeleteUser(c *gin.Context, userID uuid.UUID) error {
	if err := u.userRepo.DeleteUser(c, userID); err != nil {
		return err
	}
	return nil
}

func (u *UserCase) LoginExists(c *gin.Context, login string) (bool, error) {
	_, err := u.userRepo.GetUserIDByLogin(c, login)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (u *UserCase) Close() {
	u.userRepo.Close()
}
