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

func TestLoginBySetupRouter(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	account := JWT.LoginPayload{
		Account_Id:  "apitest",
		Account_Pwd: "1234",
	}
	payload, err := json.Marshal(&account)
	assert.NoError(t, err)

	router := SetupRouter()
	w := httptest.NewRecorder()
	request, err := http.NewRequest("POST", "/api/public/login", bytes.NewBuffer(payload))
	assert.NoError(t, err)

	router.ServeHTTP(w, request)

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
	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.POST("/api/department/:name", AddDepartment)

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
	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/department/", ReadDepartment)

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
	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/department/", ReadDepartment)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/department/?page=2&limit=1", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReadDepartmentInvalidPaging(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/department/", ReadDepartment)

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

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.PUT("/api/department/:id/:new", UpdateDepartment)

	// Create Data for Test
	newName := "New Name"
	newDepartment := Department{
		Department_Name: "Old Name",
	}
	db.Create(&newDepartment)

	w := httptest.NewRecorder()
	requrl := "/api/department/" + strconv.FormatUint(uint64(newDepartment.ID), 10) + "/" + newName
	fmt.Println(requrl)
	request, _ := http.NewRequest("PUT", requrl, nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	db.Where("id = ?", newDepartment.ID).Find(&result)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "New Name", result.Department_Name)

	db.Unscoped().Where("id = ?", newDepartment.ID).Delete(&Department{})
}

func TestDeleteDepartment(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)
	var result Department

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.DELETE("/api/department/:id", DeleteDepartment)

	test := Department{ // 지울 data 정보
		Department_Name: "deleteTest",
	}
	createResult := db.Create(&test) // 지울 data 넣기
	assert.NoError(t, createResult.Error)

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

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.DELETE("/api/department/:id", DeleteDepartment)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("DELETE", "/api/department/-1", nil) // Use invalid department id
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSearchDepartmentByName(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/department/:name", SearchDepartmentByName)

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
	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.POST("/api/employee/:name", AddEmployee)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/api/employee/Test Employee", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Test Employee", result["msg"])

	//db.Unscoped().Where("Employee_Name = ?", "Test Employee").Delete(&Employee{})
}

func TestAddEmployeeBatch(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	var result map[string]interface{}
	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.POST("/api/employee/batch/:count/:days", AddEmployeeBatch)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/api/employee/batch/150/3", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "batch create complete", result["msg"])
}

func TestReadEmployee(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	var results []Employee
	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/employee/", ReadEmployee)

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
	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/employee/", ReadEmployee)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/employee/?page=2&limit=2", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReadEmployeeInvalidPaging(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/employee/", ReadEmployee)

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

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.PUT("/api/employee/:id/:new", UpdateEmployee)

	// Create Data for Test
	newName := "New Name"
	newEmployee := Employee{
		Employee_Name: "Old Name",
	}
	db.Create(&newEmployee)

	w := httptest.NewRecorder()
	requrl := "/api/employee/" + strconv.FormatUint(uint64(newEmployee.ID), 10) + "/" + newName
	request, _ := http.NewRequest("PUT", requrl, nil) // id=1인 employee의 name을 UJS로 UPDATE
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	db.Where("id = ?", newEmployee.ID).Find(&result)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "New Name", result.Employee_Name)

	db.Unscoped().Where("id = ?", newEmployee.ID).Delete(&Employee{})
}

func TestDeleteEmployee(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)
	var result Employee

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.DELETE("/api/employee/:id", DeleteEmployee)

	test := Employee{ // 지울 data 정보
		Employee_Name: "deleteTest",
	}
	createResult := db.Create(&test) // 지울 data 넣기
	assert.NoError(t, createResult.Error)

	w := httptest.NewRecorder()
	requrl := "/api/employee/" + strconv.FormatUint(uint64(test.ID), 10)
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

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.DELETE("/api/employee/:id", DeleteEmployee)

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

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/employee/name/:name", SearchEmployeeByName)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/employee/name/Test Employee", nil)
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

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/employee/day/:days", SearchEmployeeByDay)

	days := 4
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

func TestSearchEmployeeByDayInvalidPaging(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/employee/day/:days", SearchEmployeeByDay)

	days := 4
	w := httptest.NewRecorder()
	requrl := "/api/employee/day/" + strconv.Itoa(days) + "?page=-1&limit=-1"
	request, _ := http.NewRequest("GET", requrl, nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSearchEmployeeByDayPaging(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)
	var results []Employee

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/employee/day/:days", SearchEmployeeByDay)

	days := 4
	w := httptest.NewRecorder()
	requrl := "/api/employee/day/" + strconv.Itoa(days) + "?page=1&limit=1" //하나는 data가 있을 것이라고 가정
	request, _ := http.NewRequest("GET", requrl, nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, len(results))
}

func TestAddEmployeeDepartment(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.POST("/api/assign/:eid/:did", AddEmployeeDepartment)

	// Create Test Data
	var result map[string]interface{}
	newEmployee := Employee{
		Employee_Name: "TestEmployee",
	}
	newDepartment := Department{
		Department_Name: "TestDepartment",
	}

	db.Create(&newEmployee)
	db.Create(&newDepartment)

	w := httptest.NewRecorder()
	requrl := "/api/assign/" + strconv.FormatUint(uint64(newEmployee.ID), 10) + "/" + strconv.FormatUint(uint64(newDepartment.ID), 10)
	request, _ := http.NewRequest("POST", requrl, nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, newEmployee.Employee_Name, result["employee"])
	assert.Equal(t, newDepartment.Department_Name, result["department"])

	db.Model(&newEmployee).Association("Employee_Departments").Clear() // 테스트 후 종속성 삭제
	db.Unscoped().Where("id = ?", newEmployee.ID).Delete(&Employee{})  // 테스트 후 정보 삭제
	db.Unscoped().Where("id = ?", newDepartment.ID).Delete(&Department{})
}

func TestAddEmployeeDepartmentInvalidId(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.POST("/api/assign/:eid/:did", AddEmployeeDepartment)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/api/assign/-1/-1", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteEmployeeDepartment(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.DELETE("/api/assign/:eid/:did", DeleteEmployeeDepartment)

	// Create Test Data
	var result map[string]interface{}
	newEmployee := Employee{
		Employee_Name: "TestEmployee",
	}
	newDepartment := Department{
		Department_Name: "TestDepartment",
	}

	db.Create(&newEmployee)
	db.Create(&newDepartment)
	db.Model(&newEmployee).Association("Employee_Departments").Append(&newDepartment)

	w := httptest.NewRecorder()
	requrl := "/api/assign/" + strconv.FormatUint(uint64(newEmployee.ID), 10) + "/" + strconv.FormatUint(uint64(newDepartment.ID), 10)
	request, _ := http.NewRequest("DELETE", requrl, nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "employee exited by department", result["msg"])

	db.Unscoped().Where("id = ?", newEmployee.ID).Delete(&Employee{})
	db.Unscoped().Where("id = ?", newDepartment.ID).Delete(&Department{})
}

func TestDeleteEmployeeDepartmentInvalidId(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.DELETE("/api/assign/:eid/:did", DeleteEmployeeDepartment)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("DELETE", "/api/assign/-1/-1", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestReadEmployeeInDepartment(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/assign/:did", ReadEmployeeInDepartment)

	// Create Test Data
	newEmployee := Employee{
		Employee_Name: "TestEmployee",
	}
	newDepartment := Department{
		Department_Name: "TestDepartment",
	}

	db.Create(&newEmployee)
	db.Create(&newDepartment)
	db.Model(&newEmployee).Association("Employee_Departments").Append(&newDepartment)

	w := httptest.NewRecorder()
	requrl := "/api/assign/" + strconv.FormatUint(uint64(newDepartment.ID), 10)
	request, _ := http.NewRequest("GET", requrl, nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)

	db.Model(&newEmployee).Association("Employee_Departments").Clear()
	db.Unscoped().Where("id = ?", newEmployee.ID).Delete(&Employee{})
	db.Unscoped().Where("id = ?", newDepartment.ID).Delete(&Department{})
}

func TestReadEmployeeInDepartmentInvalidId(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	token, err := JWT.GenerateToken("gotest")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/assign/:did", ReadEmployeeInDepartment)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/assign/-1", nil) // did=24000.. 인 Department가 없다고 가정
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

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
