package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

/* 사원에게 부서 만들어주기 */
func AddEmployeeDepartment(c *gin.Context) {
	eid := c.Param("eid")
	did := c.Param("did")

	var employee Employee
	var department Department

	db.Where("id = ?", eid).Find(&employee)
	db.Where("id = ?", did).Find(&department)

	if (employee.ID == 0) || (department.ID == 0) {
		log.Println("Id error at employee or department", employee.ID, department.ID)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error allocate department to employee",
		})
		c.Abort()
		return
	}

	db.Model(&employee).Association("Employee_Departments").Append(&department)
	c.JSON(http.StatusOK, gin.H{
		"employee":   employee.Employee_Name,
		"department": department.Department_Name,
	})
}

/* 사원을 부서에서 제외시키기 */
func DeleteEmployeeDepartment(c *gin.Context) {
	eid := c.Param("eid")
	did := c.Param("did")

	var employee Employee
	var department Department

	db.Where("id = ?", eid).Find(&employee)
	db.Where("id = ?", did).Find(&department)

	if (employee.ID == 0) || (department.ID == 0) {
		log.Println("id error at employee or department")

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error id at department or employee",
		})
		c.Abort()
		return
	}

	db.Model(&employee).Association("Employee_Departments").Delete(&department)
	c.JSON(http.StatusOK, gin.H{
		"msg":        "employee exited by department",
		"employee":   employee.Employee_Name,
		"department": department.Department_Name,
	})
}

/* 부서가 없는 사원을 부서로 자동 매칭 */
func AutoMatchEmployeeToDepartment(c *gin.Context) { // /api/assign/auto POST
	var employees []Employee
	var departments []Department
	var random int

	db.Find(&employees)
	db.Find(&departments)
	depCnt := len(departments)

	for i := 0; i < len(employees); i++ {
		cnt := db.Model(&employees[i]).Association("Employee_Departments").Count()
		fmt.Println(employees[i], cnt)
		if cnt == 0 {
			random = rand.Intn(depCnt)
			db.Model(&employees[i]).Association("Employee_Departments").Append(&departments[random])
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "matching finish",
	})
}
