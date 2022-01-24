package main

import (
	"github.com/gin-gonic/gin"

	Oauth "github.com/dunebi/myapi-oauth"
)

/* API 세팅 */
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// callback by oauth CA
	r.GET("/auth/callback/google", func(c *gin.Context) {
		Oauth.LoginCallback(c, db)
	})
	r.GET("/auth/callback/facebook", func(c *gin.Context) {
		Oauth.LoginCallback(c, db)
	})
	r.GET("/auth/callback/github", func(c *gin.Context) {
		Oauth.LoginCallback(c, db)
	})

	r.POST("/init", InitTable)
	r.DELETE("/delete", DeleteTable)
	r.GET("/login/*CA", func(c *gin.Context) {
		Oauth.Login(c)
	})

	// To run in Postman
	api := r.Group("/api")
	{
		// Use를 통해 Middleware인 AuthorizeAccount를 가져와 MiddleWare에서 검증 진행
		department := api.Group("/department").Use(AuthorizeAccount())
		{
			department.GET("/only", ReadDepartmentOnly)
			department.GET("/", ReadDepartment)
			department.GET("/:name", SearchDepartmentByName)
			department.GET("/:name/employee", ReadEmployeeInDepartment) // 부서에 속한 직원 명단 가져오기
			department.PUT("/", UpdateDepartment)
			department.POST("/", AddDepartment)
			department.DELETE("/:name", DeleteDepartment)
		}
		employee := api.Group("/employee").Use(AuthorizeAccount())
		{
			employee.GET("/", ReadEmployee)
			employee.GET("/name/:name", SearchEmployeeByName)
			employee.GET("/day/:days", SearchEmployeeByDay)
			employee.PUT("/:id", UpdateEmployee)
			employee.POST("/", AddEmployee)
			employee.DELETE("/:name", DeleteEmployee)
			employee.DELETE("/id/:id", DeleteEmployeById)
		}
		assign := api.Group("/assign").Use(AuthorizeAccount())
		{
			assign.POST("/:name/:department", AddEmployeeDepartment)
			assign.POST("/id/:eid/:department", AddEmployeeDepartmentById)
			assign.DELETE("/:name/:department", DeleteEmployeeDepartment)
			assign.DELETE("/id/:eid/:department", DeleteEmployeeDepartmentById)
		}
	}

	return r
}
