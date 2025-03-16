package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/lemavisaitov/lk-api/internal/apperr"
	"github.com/lemavisaitov/lk-api/internal/logger"
	"github.com/lemavisaitov/lk-api/internal/model"
	"github.com/lemavisaitov/lk-api/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

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

	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		errMessage := ""
		for _, err := range err.(validator.ValidationErrors) {
			errMessage += fmt.Sprintf("ошибка в поле %s: %s\n", err.StructField(), err.ActualTag())
		}
		logger.Error("error in signup request",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": errMessage})
		return
	}

	ok, err := h.userUC.LoginExists(c, user.Login)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login already exists"})
		return
	}

	id, err := uuid.NewV7()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.ID = id

	_, err = h.userUC.AddUser(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *Handle) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		errMessage := ""
		for _, err := range err.(validator.ValidationErrors) {
			errMessage += fmt.Sprintf("ошибка в поле %s: %s\n", err.StructField(), err.ActualTag())
		}
		logger.Error("error in login request",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": errMessage})
		return
	}

	userID, err := h.userUC.GetUserIDByLogin(c, req.Login)

	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userUC.GetUser(c, *userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if user.Password != req.Password {
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

	user, err := h.userUC.GetUser(c, id)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"name": user.Name, "age": user.Age})
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

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		errMessage := ""
		for _, err := range err.(validator.ValidationErrors) {
			errMessage += fmt.Sprintf("ошибка в поле %s: %s\n", err.StructField(), err.ActualTag())
		}
		logger.Error("error in signup request",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": errMessage})
		return
	}

	req.ID = id
	_, err = h.userUC.UpdateUser(c, req)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}
