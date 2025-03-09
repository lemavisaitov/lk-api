package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/lemavisaitov/lk-api/internal/model"
	"github.com/lemavisaitov/lk-api/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Handler interface {
	Signup(*gin.Context)
	Login(*gin.Context)
	GetUser(*gin.Context)
	UpdateUser(*gin.Context)
	DeleteUser(*gin.Context)
}

type Handle struct {
	userUC usecase.UserProvider
}

func New(userProvider usecase.UserProvider) *Handle {
	return &Handle{
		userUC: userProvider,
	}
}

func (h *Handle) Signup(c *gin.Context) {
	var user model.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if user.Password == "" || user.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password or name is empty"})
		return
	}
	ok, err := h.userUC.LoginExists(c, user.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login already exists"})
		return
	}

	if id, err := uuid.NewV7(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		user.ID = id
	}

	if id, err := h.userUC.AddUser(c, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"id": id})
	}
}

func (h *Handle) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.userUC.GetUserIDByLogin(c, req.Login)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if userID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login does not exist"})
		return
	}

	if user, err := h.userUC.GetUser(c, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if user.Password != req.Password {
		log.Println(user.Password, req.Password)
		c.JSON(http.StatusForbidden, gin.H{"error": "wrong password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": userID})
}

func (h *Handle) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user, err := h.userUC.GetUser(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if user.ID == uuid.Nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	} else {
		c.JSON(http.StatusOK, gin.H{"name": user.Name, "age": user.Age})
	}
}

func (h *Handle) UpdateUser(c *gin.Context) {
	var req model.UpdateUserRequest

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = id

	if id, err := h.userUC.UpdateUser(c, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if id == uuid.Nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "id does not exist"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *Handle) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.userUC.DeleteUser(c, id)
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
