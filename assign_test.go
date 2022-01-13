package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/dunebi/myapi/JWT"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

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
