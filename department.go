package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Department Table
type Department struct {
	ID                   uint        `gorm:"primaryKey"`
	Department_Name      string      `gorm:"unique"`
	Department_Employees []*Employee `gorm:"many2many:employee_departments"`
}

type dData struct {
	DName []string `json:"dname" binding:"required"`
}

type UpdateData struct {
	PrevName string `json:"prev" binding:"required"`
	NewName  string `json:"new" binding:"required"`
}

/* 새로운 Department를 추가(C) */
func AddDepartment(c *gin.Context) {
	var data dData
	var department Department
	msg := make([]string, 0, 3)
	var temp string
	err := c.ShouldBindJSON(&data)

	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "invalid json",
		})
		c.Abort()
		return
	}

	for i := 0; i < len(data.DName); i++ {
		department.Department_Name = data.DName[i]
		result := db.Create(&Department{Department_Name: data.DName[i]})
		if result.Error != nil {
			log.Println(result.Error)
			temp = data.DName[i] + ": Create Fail!"
			msg = append(msg, temp)
			continue
		}
		temp = data.DName[i] + ": Create Success\n"
		msg = append(msg, temp)
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": msg,
	})
}

/* Department Table 불러오기(R)_Paging 추가 */
func ReadDepartment(c *gin.Context) { // localhost:8080/api/department/?page= & limit= (GET)
	var departments []Department
	limit, page, sort := Paging(c)
	offset := (page - 1) * limit

	result := db.Limit(limit).Offset(offset).Order(sort).Preload("Department_Employees").Find(&departments)
	//result := db.Find(&departments)

	if result.Error != nil {
		log.Println(result.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "READ error",
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, departments)
}

func ReadDepartmentOnly(c *gin.Context) {
	var departments []Department
	limit, page, sort := Paging(c)
	offset := (page - 1) * limit

	result := db.Limit(limit).Offset(offset).Order(sort).Find(&departments)
	//result := db.Find(&departments)

	if result.Error != nil {
		log.Println(result.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "READ error",
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, departments)
}

/* 기존의 Department 내용 수정(U) */
func UpdateDepartment(c *gin.Context) { // localhost:8080/api/department
	var department Department
	var data UpdateData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "invalid json",
		})
		c.Abort()
		return
	}

	result := db.Model(&department).Where("Department_Name = ?", data.PrevName).Update("Department_Name", data.NewName)
	if result.Error != nil {
		log.Println(result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "UPDATE error",
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "Department Update Complete",
	})
}

/* 기존의 Department 삭제(D) */
func DeleteDepartment(c *gin.Context) {
	name := c.Param("name")

	var department Department

	// Find Department
	db.Where("Department_Name = ?", name).Find(&department)
	fmt.Println(department)
	if department.ID == 0 { // 테이블에 이름이 일치하는 Department가 없으면 ID = 0 으로 반환
		log.Println("Department name incorrect")

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error deleting department",
		})
		c.Abort()
		return
	}

	db.Model(&department).Association("Department_Employees").Clear()
	db.Delete(&department)

	c.JSON(http.StatusOK, gin.H{
		"msg": "Delete Complete",
	})
}

/* 해당 이름의 모든 부서 조회 */
func SearchDepartmentByName(c *gin.Context) {
	name := c.Param("name")

	var departments []Department

	result := db.Where("Department_Name = ?", name).Preload("Department_Employees").Find(&departments)
	if result.Error != nil {
		log.Println(result.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": result.Error.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, departments)
}

/* 부서 내 소속된 사원 목록 출력 */
func ReadEmployeeInDepartment(c *gin.Context) {
	dname := c.Param("name")

	var employees []Employee
	var department Department

	result := db.Where("Department_Name=?", dname).Find(&department)

	limit, page, sort := Paging(c)
	offset := (page - 1) * limit
	fmt.Println(department)

	if (result.Error != nil) || (department.ID == 0) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Error on Read Employees in Department",
		})
		c.Abort()
		return
	}
	log.Println(department)

	//db.Model(&department).Association("Department_Employees").Find(&employees)
	db.Limit(limit).Offset(offset).Order(sort).Model(&department).Association("Department_Employees").Find(&employees)

	c.JSON(http.StatusOK, employees)
}
