package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSPUGWSYZ")

// Employee Table
type Employee struct {
	ID                   uint      `gorm:"primaryKey"`
	EntryTime            time.Time `gorm:"autoCreateTime"`
	Employee_Name        string
	Employee_Departments []*Department `gorm:"many2many:employee_departments"`
}

type eData struct {
	EName string `json:"ename" binding:"required"`
	DName string `json:"dname"`
}

/* 새로운 Employee 추가(C) */
func AddEmployee(c *gin.Context) {
	var data []eData
	var department Department
	var employee Employee
	var temp string
	msg := make([]string, 0, 3)
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "invalid json",
		})
		c.Abort()
		return
	}

	for i := 0; i < len(data); i++ {
		employee = Employee{}
		employee.Employee_Name = data[i].EName
		if data[i].DName == "" {
			db.Create(&employee)
			temp = employee.Employee_Name + ": Create Success without department"
			msg = append(msg, temp)

		} else {
			department = Department{}
			db.Where("Department_Name = ?", data[i].DName).Find(&department)
			if department.ID == 0 { // department name incorrect. Abort API with msg
				temp = employee.Employee_Name + ": Create Fail. Department " + data[i].DName + " is not exist"
				msg = append(msg, temp)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"msg":           msg,
					"not processed": data[i:],
				})
				return
			}

			db.Create(&employee)
			db.Model(&employee).Association("Employee_Departments").Append(&department)

			temp = employee.Employee_Name + ": Create Success in department " + data[i].DName
			msg = append(msg, temp)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": msg,
	})
}

/* Employee Table 불러오기(R)_By Paging */
func ReadEmployee(c *gin.Context) {
	var employees []Employee
	limit, page, sort := Paging(c)
	offset := (page - 1) * limit

	//result := db.Find(&employees)
	result := db.Limit(limit).Offset(offset).Order(sort).
		Preload("Employee_Departments").Find(&employees)

	if result.Error != nil {
		log.Println(result.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Read Error",
		})
		c.Abort()
		return
	}
	/*
		var resultEmployees []readEmployee
		for i := 0; i < len(employees); i++ {
			tmp := readEmployee{employees[i].ID, employees[i].Employee_Name, nil}
			for j := 0; j < len(employees[i].Employee_Departments); j++ {
				tmp.Department = append(tmp.Department, employees[i].Employee_Departments[j].Department_Name)
			}
			resultEmployees = append(resultEmployees, tmp)
		}
	*/

	c.JSON(http.StatusOK, employees)
}

/* 기존의 Employee 내용 수정(U) */
func UpdateEmployee(c *gin.Context) {
	var employee Employee
	dataId := c.Param("id")

	var data eData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "invalid json",
		})
		c.Abort()
		return
	}

	result := db.Model(&employee).Where("id = ?", dataId).Update("Employee_Name", data.EName)
	if result.Error != nil {
		log.Println(result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Update error",
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "Employee Update Complete",
	})
}

/* 기존의 Emplpyee 삭제(D) */
func DeleteEmployee(c *gin.Context) {
	eName := c.Param("name")

	var employees []Employee
	db.Where("Employee_Name = ?", employees).Find(&employees)
	db.Where("Employee_Name = ?", eName).Find(&employees)
	if len(employees) > 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"employee info": employees,
			"msg":           "There're employees with same name",
			"can use":       "/api/employee/id/:id",
		})
		c.Abort()
		return
	} else if len(employees) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "No such employee",
		})
		c.Abort()
		return
	}

	db.Model(&employees[0]).Association("Employee_Departments").Clear()
	db.Delete(&employees[0])
	c.JSON(http.StatusOK, gin.H{
		"msg": "Delete Complete",
	})
}

func DeleteEmployeById(c *gin.Context) {
	employee_id := c.Param("id")

	var employee Employee

	// Find employee
	db.Where("id=?", employee_id).Find(&employee)
	if employee.ID == 0 {
		log.Println("Employee id incorrect")

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error deleting employee",
		})
		c.Abort()
		return
	}

	db.Model(&employee).Association("Employee_Departments").Clear()
	db.Delete(&employee)
	c.JSON(http.StatusOK, gin.H{
		"msg": "Delete Complete",
	})
}

/* n일 이내 입사한 사원 조회_Paging 추가 */
func SearchEmployeeByDay(c *gin.Context) {
	n := c.Param("days")
	var employees []Employee
	limit, page, sort := Paging(c)
	offset := (page - 1) * limit

	//result := db.Where("TO_DAYS(SYSDATE()) - TO_DAYS(created_at) <= ?", n).Find(&employees)
	result := db.Limit(limit).Offset(offset).Order(sort).Where(
		"TO_DAYS(SYSDATE()) - TO_DAYS(entry_time) <= ?", n).Preload("Employee_Departments").Find(&employees)
	if result.Error != nil {
		log.Println(result.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": result.Error.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, employees)
}

/* 해당 이름의 모든 사원 조회 */
func SearchEmployeeByName(c *gin.Context) {
	name := c.Param("name")

	var employees []Employee

	result := db.Where("Employee_Name = ?", name).Preload("Employee_Departments").Find(&employees)
	if result.Error != nil {
		log.Println(result.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": result.Error.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, employees)
}

/* Random String으로 이름 이니셜 생성 */
func RandString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
