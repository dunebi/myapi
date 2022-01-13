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
