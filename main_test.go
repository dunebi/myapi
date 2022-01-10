package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/dunebi/myapi/JWT"
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

	Login(c)
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

	Login(c)
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

	Login(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthorizeAccountNoHeader(t *testing.T) {
	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/protected/profile", Profile)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/protected/profile", nil)
	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuthorizeAccountInvalidTokenFormat(t *testing.T) {
	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/protected/profile", Profile)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/protected/profile", nil)
	request.Header.Add("Authorization", "invalid format")

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthorizeAccountInvalidToken(t *testing.T) {
	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/protected/profile", Profile)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/protected.profile", nil)
	request.Header.Add("Authorization", "Bearer 12341rnjsdkngkj3m2oithfrsd")

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthorizeAccount(t *testing.T) {
	var result Account
	err = InitDB()
	assert.NoError(t, err)

	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/protected/profile", Profile)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/protected/profile", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gotest", result.Account_Id)

	db.Unscoped().Where("id = ?", result.ID).Delete(&Account{}) // temp 정보 삭제를 위해 맨 뒤에 함수 배치
}

/* Department CRUD Test */

func TestAddDepartment(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	var result map[string]interface{}
	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/api/department/Test Department", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Test Department", result["msg"])

	db.Unscoped().Where("Department_Name = ?", "Test Department").Delete(&Department{})
}

func TestReadDepartment(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	var results []Department
	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/department/", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	//fmt.Println(results)
}

func TestReadDepartmentPaging(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	var results []Department
	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/department/?page=2&limit=1", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, len(results)) // page가 2 이므로, limit가 1이므로 -> 2번째 data가 출력되는 상황
	//fmt.Println(results)
}

func TestReadDepartmentInvalidPaging(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/department/?page=-1&limit=-1", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateDepartment(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)
	var result Department

	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("PUT", "/api/department/1/A", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	db.Where("id = ?", "1").Find(&result)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "A", result.Department_Name)
}

func TestDeleteDepartment(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)
	var result Department

	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	test := Department{ // 지울 data 정보
		Department_Name: "deleteTest",
	}
	createResult := db.Create(&test) // 지울 data 넣기
	assert.NoError(t, createResult.Error)

	router := SetupRouter()
	w := httptest.NewRecorder()
	requrl := "/api/department/" + strconv.FormatUint(uint64(test.ID), 10)
	request, _ := http.NewRequest("DELETE", requrl, nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	db.Where("id=?", test.ID).Find(&result)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint(0), result.ID) // 삭제 후에는 id = 0 으로 조회됨
	//fmt.Println(result)
}

func TestDeleteDepartmentInvalidDid(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("DELETE", "/api/department/-1", nil) // Use invalid department id
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSearchDepartmentByName(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/department/A", nil) // Search Department which name is 'A'
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
}

/* Employee Table CRUD Test */

func TestAddEmployee(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	var result map[string]interface{}
	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/api/employee/Test Employee", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Test Employee", result["msg"])

	db.Unscoped().Where("Employee_Name = ?", "Test Employee").Delete(&Employee{})
}

func TestReadEmployee(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	var results []Employee
	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/employee/", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	//fmt.Println(results)
}

func TestReadEmployeePaging(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	var results []Employee
	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/employee/?page=2&limit=5", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 5, len(results))
	//fmt.Println(results)
}

func TestReadEmployeeInvalidPaging(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/employee/?page=-1&limit=-1", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateEmployee(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)
	var result Employee

	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("PUT", "/api/employee/1/UJS", nil) // id=1인 employee의 name을 UJS로 UPDATE
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	db.Where("id = ?", "1").Find(&result)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "UJS", result.Employee_Name)
}

func TestDeleteEmployee(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)
	var result Employee

	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	test := Employee{ // 지울 data 정보
		Employee_Name: "deleteTest",
	}
	createResult := db.Create(&test) // 지울 data 넣기
	assert.NoError(t, createResult.Error)

	router := SetupRouter()
	w := httptest.NewRecorder()
	requrl := "/api/employee/" + strconv.FormatUint(uint64(test.ID), 10)
	//fmt.Println(requrl)
	request, _ := http.NewRequest("DELETE", requrl, nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	db.Where("id=?", test.ID).Find(&result)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint(0), result.ID) // 삭제 후에는 id = 0 으로 조회됨
}

func TestDeleteEmployeeInvalidEid(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("DELETE", "/api/employee/-1", nil) // Use invalid department id
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSearchEmployeeByName(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)
	var results []Employee

	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/employee/name/IJY", nil) // Search Department which name is 'A'
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	//fmt.Println(results)
}

func TestSearchEmployeeByDay(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)
	var results []Employee

	account := Account{
		Account_Id:  "gotest",
		Account_Pwd: "1234",
	}
	token, err := JWT.GenerateToken(account.Account_Id)
	assert.NoError(t, err)

	pwd, err := HashPassword(account.Account_Pwd)
	assert.NoError(t, err)
	account.Account_Pwd = pwd

	days := 4
	router := SetupRouter()
	w := httptest.NewRecorder()
	requrl := "/api/employee/day/" + strconv.Itoa(days)
	request, _ := http.NewRequest("GET", requrl, nil) // Search Department which name is 'A'
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	//fmt.Println(results)
}
