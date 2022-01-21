package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dunebi/myapi/JWT"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

/*
func TestCheckPassword(t *testing.T) {
	pwd, err := HashPassword("password")
	assert.NoError(t, err)

	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: pwd,
	}
	err = account.CheckPassword("password")
	assert.NoError(t, err)
}

func TestCheckPasswordInvalid(t *testing.T) {
	pwd, err := HashPassword("password")
	assert.NoError(t, err)

	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: pwd,
	}
	err = account.CheckPassword("passwordss")
	assert.Equal(t, "crypto/bcrypt: hashedPassword is not the hash of the given password", err.Error())
}

func TestRegister(t *testing.T) {
	var result Account
	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}

	payload, err := json.Marshal(&account)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/public/register/", bytes.NewBuffer(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = request

	err = InitDB() // 전역변수 db를 여기서 초기화해야 main.go 코드의 db변수가 문제없이 작동(없으면 runtime error가 났음)
	assert.NoError(t, err)

	Register(c) // 회원가입 후 JSON을 반환
	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, account.Account_Id, result.Account_Id)
}

func TestRegisterInvalidJSON(t *testing.T) {
	account := "gotest"

	payload, err := json.Marshal(&account)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/public/register/", bytes.NewBuffer(payload))
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = request
	Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
*/

func TestLogin(t *testing.T) {
	account := JWT.LoginPayload{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}

	payload, err := json.Marshal(&account)
	assert.NoError(t, err)
	request, err := http.NewRequest("POST", "/api/public/login/", bytes.NewBuffer(payload))
	assert.NoError(t, err)

	err = InitDB()
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = request

	LoginGoogle(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoginInvalidJSON(t *testing.T) {
	account := "gotest"

	payload, err := json.Marshal(&account)
	assert.NoError(t, err)
	request, err := http.NewRequest("POST", "/api/public/login/", bytes.NewBuffer(payload))
	assert.NoError(t, err)

	err = InitDB()
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = request

	LoginGoogle(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginInvalidPassword(t *testing.T) {
	account := JWT.LoginPayload{
		Account_Id:  "gotest",
		Account_Pwd: "4321",
	}

	payload, err := json.Marshal(&account)
	assert.NoError(t, err)
	request, err := http.NewRequest("POST", "/api/public/login/", bytes.NewBuffer(payload))
	assert.NoError(t, err)

	err = InitDB()
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = request

	LoginGoogle(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
