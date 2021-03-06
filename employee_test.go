package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

/* Employee Table CRUD Test */

func TestAddEmployee(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	var result map[string]interface{}
	token, err := GenerateToken("gotest", "myCA")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.POST("/api/employee/:name/:department", AddEmployee)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/api/employee/Test Employee/Test Department", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)

	db.Where("Employee_Name = ?", "Test Employee").Association("Employee_Departments").Clear()
	db.Unscoped().Where("Employee_Name=?", "Test Employee").Delete(&Employee{})
	db.Unscoped().Where("Department_Name=?", "Test Department").Delete(&Department{})
}

func TestReadEmployee(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	var results []Employee
	token, err := GenerateToken("gotest", "myCA")
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
	token, err := GenerateToken("gotest", "myCA")
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

	token, err := GenerateToken("gotest", "myCA")
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

	token, err := GenerateToken("gotest", "myCA")
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
	request, _ := http.NewRequest("PUT", requrl, nil) // id=1??? employee??? name??? UJS??? UPDATE
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

	token, err := GenerateToken("gotest", "myCA")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.DELETE("/api/employee/:id", DeleteEmployee)

	test := Employee{ // ?????? data ??????
		Employee_Name: "deleteTest",
	}
	createResult := db.Create(&test) // ?????? data ??????
	assert.NoError(t, createResult.Error)

	w := httptest.NewRecorder()
	requrl := "/api/employee/" + strconv.FormatUint(uint64(test.ID), 10)
	request, _ := http.NewRequest("DELETE", requrl, nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	db.Where("id=?", test.ID).Find(&result)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint(0), result.ID) // ?????? ????????? id = 0 ?????? ?????????
}

func TestDeleteEmployeeInvalidEid(t *testing.T) {
	err = InitDB()
	assert.NoError(t, err)

	token, err := GenerateToken("gotest", "myCA")
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

	token, err := GenerateToken("gotest", "myCA")
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

	token, err := GenerateToken("gotest", "myCA")
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

	token, err := GenerateToken("gotest", "myCA")
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

	token, err := GenerateToken("gotest", "myCA")
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(AuthorizeAccount())
	router.GET("/api/employee/day/:days", SearchEmployeeByDay)

	days := 4
	w := httptest.NewRecorder()
	requrl := "/api/employee/day/" + strconv.Itoa(days) + "?page=1&limit=1" //????????? data??? ?????? ???????????? ??????
	request, _ := http.NewRequest("GET", requrl, nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, request)

	err = json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, len(results))
}
