package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

func signup() (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"login": "admin", "password": "secret", "name": "lema", "age": 19}`

	c.Request = httptest.NewRequest("POST", "/user/signup", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	return w, c
}

func getuser(id string) (*httptest.ResponseRecorder, *gin.Context) {
	getW := httptest.NewRecorder()
	getC, _ := gin.CreateTestContext(getW)

	getC.Params = []gin.Param{{Key: "id", Value: id}}
	getC.Request = httptest.NewRequest("GET", "/user/"+id, nil)
	getC.Request.Header.Set("Accept", "application/json")

	return getW, getC
}

func TestSignup(t *testing.T) {
	w, c := signup()

	service := NewService()
	service.signup(c)

	if w.Code != 200 {
		t.Errorf("Want 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
	}

	if _, ok := response["id"]; !ok {
		t.Error("Expected 'id' in response")
	}
}

func TestLogin(t *testing.T) {
	_, signupCtx := signup()

	service := NewService()
	service.signup(signupCtx)

	loginW := httptest.NewRecorder()
	loginCtx, _ := gin.CreateTestContext(loginW)

	body := `{"login": "admin", "password": "secret"}`
	loginCtx.Request = httptest.NewRequest("POST", "/user/login", strings.NewReader(body))
	loginCtx.Request.Header.Set("Content-Type", "application/json")

	service.login(loginCtx)

	if loginW.Code != 200 {
		t.Errorf("Want 200, got %d", loginW.Code)
	}

	var response map[string]interface{}

	err := json.Unmarshal(loginW.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
	}

	if _, ok := response["id"]; !ok {
		t.Error("Expected 'id' in response")
	}
}

func TestGetUser(t *testing.T) {
	signupW, signupC := signup()

	service := NewService()
	service.signup(signupC)

	signupResponse := map[string]interface{}{}
	err := json.Unmarshal(signupW.Body.Bytes(), &signupResponse)
	if err != nil {
		t.Error(err)
	}

	id, ok := signupResponse["id"].(string)
	if !ok {
		t.Error("Expected 'id' in response")
	}

	getW, getCtx := getuser(id)

	service.getUser(getCtx)

	if getW.Code != 200 {
		t.Errorf("Want 200, got %d", getW.Code)
	}

	getResponse := map[string]interface{}{}
	err = json.Unmarshal(getW.Body.Bytes(), &getResponse)
	if err != nil {
		t.Error(err)
	}

	if _, ok := getResponse["name"]; !ok {
		t.Error("Expected 'name' in response")
	}
	if _, ok := getResponse["age"]; !ok {
		t.Error("Expected 'age' in response")
	}
}

func TestPut(t *testing.T) {
	signupW, signupCtx := signup()

	service := NewService()
	service.signup(signupCtx)

	signupResponse := map[string]interface{}{}
	err := json.Unmarshal(signupW.Body.Bytes(), &signupResponse)
	if err != nil {
		t.Error(err)
	}

	id, ok := signupResponse["id"].(string)
	if !ok {
		t.Error("Expected 'id' in response")
	}

	putW := httptest.NewRecorder()
	putCtx, _ := gin.CreateTestContext(putW)

	body := `{"name": "lemas", "age": 20}`
	putCtx.Params = []gin.Param{{Key: "id", Value: id}}
	putCtx.Request = httptest.NewRequest("PUT", "/user/"+id, strings.NewReader(body))
	putCtx.Request.Header.Set("Content-Type", "application/json")

	service.updateUser(putCtx)

	if putW.Code != 200 {
		t.Errorf("Want 200, got %d", putW.Code)
	}

	getW, getCtx := getuser(id)

	service.getUser(getCtx)

	if getW.Code != 200 {
		t.Errorf("Want 200, got %d", getW.Code)
	}

	var updateResponse map[string]interface{}
	err = json.Unmarshal(getW.Body.Bytes(), &updateResponse)
	if err != nil {
		t.Error(err)
	}

	if updateResponse["name"] != "lemas" || updateResponse["age"] != 20.0 {
		t.Errorf("Update failed: name = %v, age = %v", updateResponse["name"], updateResponse["age"])
	}
}

func TestDelete(t *testing.T) {
	signupW, signupCtx := signup()

	service := NewService()
	service.signup(signupCtx)

	signupResponse := map[string]interface{}{}
	err := json.Unmarshal(signupW.Body.Bytes(), &signupResponse)
	if err != nil {
		t.Error(err)
	}
	id, ok := signupResponse["id"].(string)
	if !ok {
		t.Error("Expected 'id' in response")
	}

	deleteW := httptest.NewRecorder()
	deleteCtx, _ := gin.CreateTestContext(deleteW)

	deleteCtx.Params = []gin.Param{{Key: "id", Value: id}}
	deleteCtx.Request = httptest.NewRequest("DELETE", "/user/"+id, nil)

	service.deleteUser(deleteCtx)

	if deleteW.Code != 200 {
		t.Errorf("Want 200, got %d", deleteW.Code)
	}

	getW, getCtx := getuser(id)
	service.getUser(getCtx)

	if getW.Code != 404 {
		t.Errorf("Want 404, got %d", getW.Code)
	}
}
