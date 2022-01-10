package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

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

	err = InitDB()
	assert.NoError(t, err)

	Register(c) // 회원가입 후 JSON을 반환
	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, account.Account_Id, result.Account_Id)
}

func TestInvalidRegister(t *testing.T) {
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
