package handler

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lemavisaitov/lk-api/internal/model"
	"net/http"
)

type Storage interface {
	UpdateUser(ctx context.Context, request model.UpdateUserRequest) (uuid.NullUUID, error)
	AddUser(ctx context.Context, usr model.User) error
	GetUser(ctx context.Context, id uuid.UUID) (model.User, error)
	GetUserIDByLogin(ctx context.Context, login string) (uuid.NullUUID, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type Handle struct {
	storage Storage
}

func New(storage Storage) *Handle {
	return &Handle{
		storage: storage,
	}
}

func (h *Handle) Signup(c *gin.Context) {
	var usr model.User

	if err := c.ShouldBindJSON(&usr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if usr.Password == "" || usr.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password or name is empty"})
		return
	}
	if !h.loginExists(c, usr.Login) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login already exists"})
		return
	}

	id, err := uuid.NewV7()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	usr.ID.UUID = id
	err = h.storage.AddUser(c, usr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *Handle) Login(c *gin.Context) {
	login := c.Param("login")
	password := c.Param("password")

	userID, err := h.storage.GetUserIDByLogin(c, login)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
		return
	}
	if !userID.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login does not exist"})
		return
	}

	user, err := h.storage.GetUser(c, userID.UUID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if user.Password != password {
		c.JSON(http.StatusForbidden, gin.H{"error": "wrong password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": userID.UUID})
}

func (h *Handle) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.storage.GetUser(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"name": user.Name, "age": user.Age})
}

func (h *Handle) UpdateUser(c *gin.Context) {
	var req model.UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := h.storage.UpdateUser(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !id.Valid {
		c.JSON(http.StatusNotFound, gin.H{"error": "id does not exist"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id.UUID})
}

func (h *Handle) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.storage.DeleteUser(c, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user does not exist"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *Handle) loginExists(c *gin.Context, login string) bool {
	id, err := h.storage.GetUserIDByLogin(c, login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return false
	}
	return id.Valid
}
