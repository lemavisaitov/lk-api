package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"net/http"
	"sync"
)

type user struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
	mu       sync.RWMutex
}

var userMap = make(map[uuid.UUID]*user)
var userLoginMap = make(map[string]uuid.UUID)

func main() {
	router := gin.Default()

	router.POST("/user/signup", signup)
	router.GET("user/:id", getUser)
	router.POST("/user/login", login)
	router.PUT("/user/:id", updateUser)
	router.DELETE("/user/:id", deleteUser)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}

}

func signup(c *gin.Context) {
	var usr *user
	if err := c.ShouldBindJSON(&usr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, ok := userLoginMap[usr.Login]; ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Such a login already exists"})
		return
	}
	id := uuid.New()
	userLoginMap[usr.Login] = id
	userMap[id] = usr
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func getUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	usr, ok := userMap[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	usr.mu.RLock()
	c.JSON(http.StatusOK, gin.H{"name": usr.Name, "age": usr.Age, "login": usr.Login})
	usr.mu.RUnlock()
}

func updateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrap(err, "Incorrect user uuid")})
		return
	}

	var input *user
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usr, ok := userMap[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	usr.mu.Lock()
	defer usr.mu.Unlock()
	updateParams(usr, input)
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func login(c *gin.Context) {
	var input *user
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, ok := userLoginMap[input.Login]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login does not exist"})
		return
	}

	usr := userMap[id]
	usr.mu.RLock()
	if usr.Password != input.Password {
		c.JSON(http.StatusForbidden, gin.H{"error": "Wrong password"})
		return
	}
	usr.mu.RUnlock()
	c.JSON(http.StatusOK, gin.H{"id": userLoginMap[input.Login]})
}

func deleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrap(err, "Incorrect user uuid")})
		return
	}

	usr, ok := userMap[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	usr.mu.Lock()
	defer usr.mu.Unlock()
	delete(userLoginMap, usr.Login)
	delete(userMap, id)
}

func updateParams(usr, input *user) {
	if input.Login != "" {
		usr.Login = input.Login
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
