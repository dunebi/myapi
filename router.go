package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/* API 세팅 */
func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/hi", func(c *gin.Context) {
		c.String(http.StatusOK, "hello")
	})

	// To run in Postman
	api := r.Group("/api")
	{
		public := api.Group("/public")
		{
			public.POST("/init", InitTable)
			public.DELETE("/delete", DeleteTable)
			public.POST("/login", Login)
			public.POST("/register", Register)

		}

		// Use를 통해 Middleware인 AuthorizeAccount를 가져와 MiddleWare에서 검증 진행
		department := api.Group("/department").Use(AuthorizeAccount())
		{
			department.GET("/only", ReadDepartmentOnly)
			department.GET("/", ReadDepartment)
			department.GET("/:name", SearchDepartmentByName)
			department.GET("/:name/employee", ReadEmployeeInDepartment) // 부서에 속한 직원 명단 가져오기
			department.PUT("/:id/:new", UpdateDepartment)
			department.POST("/:name", AddDepartment)
			department.DELETE("/:name", DeleteDepartment)
		}
		employee := api.Group("/employee").Use(AuthorizeAccount())
		{
			employee.GET("/", ReadEmployee)
			employee.GET("/name/:name", SearchEmployeeByName)
			employee.GET("/day/:days", SearchEmployeeByDay)
			employee.PUT("/:id/:new", UpdateEmployee)
			employee.POST("/:name/:department", AddEmployee)
			employee.POST("/:name/:department/new", AddNewDepartment)
			employee.POST("/batch/:count/:days", AddEmployeeBatch)
			employee.DELETE("/:id", DeleteEmployee)
		}
		assign := api.Group("/assign").Use(AuthorizeAccount())
		{
			assign.DELETE("/:eid/:did", DeleteEmployeeDepartment)
			assign.POST("/:eid/:did", AddEmployeeDepartment)
			assign.POST("/auto", AutoMatchEmployeeToDepartment) // 부서 없는 직원에게 부서 자동매칭
		}
	}

	return r
}
