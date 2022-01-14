package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

/* 기존 사원에게 부서 추가(이름을 Param으로 받아옴) */
func AddEmployeeDepartment(c *gin.Context) {
	eName := c.Param("name")
	dName := c.Param("department")

	var employees []Employee
	db.Where("Employee_Name = ?", eName).Preload("Employee_Departments").Find(&employees)
	if len(employees) > 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":     "There're employees with same name",
			"can use": "/api/assign/:id/:department",
			"data":    employees,
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

	var department Department
	db.Where("Department_Name = ?", dName).Find(&department)
	if department.ID == 0 {
		if dName == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": "No Department Name",
			})
			c.Abort()
			return
		}
		department.Department_Name = dName

		c.JSON(http.StatusOK, gin.H{
			"msg": "new department created",
		})
	}

	db.Model(&employees[0]).Association("Employee_Departments").Append(&department)

	c.JSON(http.StatusOK, gin.H{
		"employee":   eName,
		"department": dName,
	})
}

/* 사원에게 부서 만들어주기(ID) */
func AddEmployeeDepartmentById(c *gin.Context) {
	eid := c.Param("eid")
	department_name := c.Param("department")

	var employee Employee
	var department Department

	db.Where("id = ?", eid).Find(&employee)
	db.Where("Department_Name = ?", department_name).Find(&department)

	if employee.ID == 0 {
		log.Println("Id error at employee", employee.ID, department.ID)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error allocate department to employee",
		})
		c.Abort()
		return
	}

	if department.ID == 0 {
		if department_name == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": "No Department Name",
			})
			c.Abort()
			return
		}
		department.Department_Name = department_name

		c.JSON(http.StatusOK, gin.H{
			"msg": "new department created with new employee",
		})
	}

	db.Model(&employee).Association("Employee_Departments").Append(&department)
	c.JSON(http.StatusOK, gin.H{
		"employee":   employee.Employee_Name,
		"department": department.Department_Name,
	})
}

func DeleteEmployeeDepartment(c *gin.Context) {
	eName := c.Param("name")
	dName := c.Param("department")

	var employees []Employee
	db.Where("Employee_Name = ?", eName).Preload("Employee_Departments").Find(&employees)
	if len(employees) > 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":     "There're employees with same name",
			"can use": "/api/assign/:id/:department",
			"data":    employees,
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

	var departments []Department
	db.Model(&employees[0]).Association("Employee_Departments").Find(&departments)

	for i := 0; i < len(departments); i++ {
		if departments[i].Department_Name == dName {
			db.Model(&employees[0]).Association("Employee_Departments").Delete(&departments[i])
			c.JSON(http.StatusOK, gin.H{
				"msg":        "employee exited by department",
				"employee":   eName,
				"department": dName,
			})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"msg": "This Employee is not in such Department or No such department",
	})
	c.Abort()
}

/* 사원을 부서에서 제외시키기(ID) */
func DeleteEmployeeDepartmentById(c *gin.Context) {
	eid := c.Param("eid")
	dName := c.Param("department")

	var employee Employee
	db.Where("id = ?", eid).Find(&employee)

	if employee.ID == 0 {
		log.Println("Id error at employee", employee.ID)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error input on department or employee",
		})
		c.Abort()
		return
	}

	var departments []Department
	db.Model(&employee).Association("Employee_Departments").Find(&departments)

	for i := 0; i < len(departments); i++ {
		if departments[i].Department_Name == dName {
			db.Model(&employee).Association("Employee_Departments").Delete(&departments[i])
			c.JSON(http.StatusOK, gin.H{
				"msg":        "employee exited by department",
				"employee":   employee.Employee_Name,
				"department": dName,
			})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"msg": "This Employee is not in such Department or No such department",
	})
	c.Abort()

}
