package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"sync"
)

type user struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
}

type UserStorage struct {
	users      map[uuid.UUID]user
	userLogins map[string]uuid.UUID
	mu         sync.RWMutex
}

type Service struct {
	storage *UserStorage
}

const errUserNotFound = "user not found"

func main() {
	service := NewService()

	router := gin.Default()
	router.POST("/user/signup", service.signup)
	router.GET("user/:id", service.getUser)
	router.POST("/user/login", service.login)
	router.PUT("/user/:id", service.updateUser)
	router.DELETE("/user/:id", service.deleteUser)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}

}

func (s *Service) signup(c *gin.Context) {
	var usr user
	if err := c.ShouldBindJSON(&usr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if usr.Password == "" || usr.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password or name is empty"})
		return
	}
	if _, ok := s.storage.userLogins[usr.Login]; ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login already exists"})
		return
	}
	id := uuid.New()

	s.storage.mu.Lock()
	s.storage.userLogins[usr.Login] = id
	s.storage.users[id] = usr
	s.storage.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (s *Service) getUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usr, ok := s.storage.users[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": errUserNotFound})
		return
	}
	s.storage.mu.RLock()
	defer s.storage.mu.RUnlock()
	c.JSON(http.StatusOK, gin.H{"name": usr.Name, "age": usr.Age, "login": usr.Login})
}

func (s *Service) updateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var input user
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usr, ok := s.storage.users[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": errUserNotFound})
		return
	}

	s.storage.mu.Lock()
	s.updateParams(&usr, &input)
	s.storage.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (s *Service) login(c *gin.Context) {
	var input *user
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, ok := s.storage.userLogins[input.Login]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login does not exist"})
		return
	}

	s.storage.mu.RLock()
	usr := s.storage.users[id]
	if usr.Password != input.Password {
		c.JSON(http.StatusForbidden, gin.H{"error": "wrong password"})
		return
	}
	s.storage.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{"id": s.storage.userLogins[input.Login]})
}

func (s *Service) deleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usr, ok := s.storage.users[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": errUserNotFound})
		return
	}

	s.storage.mu.Lock()
	delete(s.storage.userLogins, usr.Login)
	delete(s.storage.users, id)
	s.storage.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func (s *Service) updateParams(usr, input *user) {
	if input.Login != "" {
		id := s.storage.userLogins[usr.Login]
		delete(s.storage.userLogins, usr.Login)
		usr.Login = input.Login
		s.storage.userLogins[usr.Login] = id
	}
	if input.Password != "" {
		usr.Password = input.Password
	}
	if input.Name != "" {
		usr.Name = input.Name
	}
	if input.Age != 0 {
		usr.Age = input.Age
	}
}

func NewService() *Service {
	return &Service{
		storage: &UserStorage{
			users:      make(map[uuid.UUID]user),
			userLogins: make(map[string]uuid.UUID),
		},
	}
}
