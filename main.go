package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dunebi/myapi/JWT"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Account Table
type Account struct {
	gorm.Model
	Account_Id  string
	Account_Pwd string
}

// Department Table
type Department struct {
	gorm.Model
	Department_Name      string
	Department_Employees []*Employee `gorm:"many2many:employee_departments;"`
}

// Employee Table
type Employee struct {
	gorm.Model
	Employee_Name        string
	Employee_Departments []*Department `gorm:"many2many:employee_departments;"`
}

var db *gorm.DB

var err error
var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSPUGWSYZ")

func main() {
	err := InitDB()
	if err != nil {
		panic("DB init error")
	}

	//db.AutoMigrate(&Account{}, &Department{}, &Employee{}) // DB Table 생성
	//db.Migrator().DropTable(&Account{}, &Department{}, &Employee{}, "employee_departments") // DB Table 삭제

	// API Server Open
	r := SetupRouter()

	r.Run(":8080")
}

/* DB를 생성(test를 위함) */
func InitDB() (err error) {
	dsn := "root:1234@tcp(127.0.0.1:3306)/myapi?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return
	}

	return
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

/* Account Table 에 데이터추가 */
func MakeNewAccount(db *gorm.DB, id string, pwd string) {
	db.Create(&Account{Account_Id: id, Account_Pwd: pwd})
}

/* 암호 해시화 */
func HashPassword(pwd string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pwd), 14)
	if err != nil {
		return "", errors.New("error") // 에러 반환
	}

	Hashed_pwd := string(bytes)
	return Hashed_pwd, nil // 문제가 없으면 error 부분에 nil을 반환
}

/* 암호 확인 (Account의 메소드) */
func (account *Account) CheckPassword(pwd string) error {
	err := bcrypt.CompareHashAndPassword([]byte(account.Account_Pwd), []byte(pwd))

	if err != nil {
		return err
	}
	return nil
}

/* API 세팅 */
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// To run in Postman
	api := r.Group("/api")
	{
		public := api.Group("/public")
		{
			public.POST("/login", Login)
			public.POST("/register", Register)
		}

		// Use를 통해 Middleware인 AuthorizeAccount를 가져와 MiddleWare에서 검증 진행
		department := api.Group("/department").Use(AuthorizeAccount())
		{
			department.GET("/", ReadDepartment)
			department.GET("/:name", SearchDepartmentByName)
			department.PUT("/:id/:new", UpdateDepartment)
			department.POST("/:name", AddDepartment)
			department.DELETE("/:id", DeleteDepartment)
		}
		employee := api.Group("/employee").Use(AuthorizeAccount())
		{
			employee.GET("/", ReadEmployee)
			employee.GET("/name/:name", SearchEmployeeByName)
			employee.GET("/day/:days", SearchEmployeeByDay)
			employee.PUT("/:id/:new", UpdateEmployee)
			employee.POST("/:name", AddEmployee)
			employee.DELETE("/:id", DeleteEmployee)
		}
		assign := api.Group("/assign").Use(AuthorizeAccount())
		{
			assign.DELETE("/:eid/:did", DeleteEmployeeDepartment)
			assign.POST("/:eid/:did", AddEmployeeDepartment)
			assign.GET("/:did", ReadEmployeeInDepartment) // 부서에 속한 직원 명단 가져오기
		}
	}

	return r
}

/* 회원가입 */
func Register(c *gin.Context) { // gin.Context를 사용. Handler 인 것으로 보임
	var account Account
	err := c.ShouldBindJSON(&account) // Context c의 내용을 JSON으로 바인딩해서 account에 넣는 것으로 보임

	if err != nil {
		log.Println(err) // fmt.Println과 달리 "log" 로 출력되어 조금 더 정보를 담고 있음

		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "invalid json",
		})
		c.Abort() // exit()와 비슷. 비정상 종료를 야기시킴

		return
	}

	// pwd hashing
	pwd, err := HashPassword(account.Account_Pwd)
	if err != nil {
		log.Println(err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error hashing password",
		})
		c.Abort()

		return
	}
	account.Account_Pwd = pwd // 받아온 account의 객체를 hash화해서 저장(불필요할수도? 또는 짜기 나름?)

	// Insert Account data to db(Account Table이 이미 automigrate됐다고 가정)
	result := db.Create(&Account{Account_Id: account.Account_Id, Account_Pwd: account.Account_Pwd})
	if result.Error != nil {
		log.Println(result.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error createing account",
		})
		c.Abort()

		return
	}

	c.JSON(http.StatusOK, account)

}

/* 로그인 */
func Login(c *gin.Context) {
	var payload JWT.LoginPayload
	var account Account

	err := c.ShouldBindJSON(&payload) // payload에 로그인 정보(Id, Pwd) 주입
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "invalid json",
		})
		c.Abort()
		return
	}

	// DB에 이 ID가 있는지 검사
	result := db.Where("Account_Id=?", payload.Account_Id).First(&account)
	if result.Error == gorm.ErrRecordNotFound { // RecordNotFound Error
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "invalid account id",
		})
		c.Abort()
		return
	}

	// password 검사
	err = account.CheckPassword(payload.Account_Pwd)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "invalid account pwd",
		})
		c.Abort()
		return
	}

	// JWT 토큰 발행
	signedToken, err := JWT.GenerateToken(account.Account_Id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error signing token",
		})
		c.Abort()
		return
	}

	tokenResponse := JWT.LoginResponse{ // string인 LoginResponse를 JSON으로 반환하기 위함
		Token: signedToken,
	}
	//fmt.Println(tokenResponse) // 형태 확인용
	c.JSON(http.StatusOK, tokenResponse)
}

/* 토큰 인증된 사용자(일명 관리자)가 사용할 API Handler들(CRUD) */

/* 새로운 Department를 추가(C) */
func AddDepartment(c *gin.Context) {
	department_name := c.Param("name")

	result := db.Create(&Department{Department_Name: department_name})
	if result.Error != nil {
		log.Println(result.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error creating department",
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": department_name,
	})
}

/* Department Table 불러오기(R) */
func ReadDepartment(c *gin.Context) { // localhost:8080/api/department/?page= & limit= (GET)
	var departments []Department
	limit, page, sort := Paging(c)
	offset := (page - 1) * limit

	result := db.Limit(limit).Offset(offset).Order(sort).Find(&departments)
	//result := db.Find(&departments)

	if result.Error != nil {
		log.Println(result.Error)

		c.String(http.StatusInternalServerError, "READ error")
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, departments)
}

/* 기존의 Department 내용 수정(U) */
func UpdateDepartment(c *gin.Context) { // localhost:8080/api/department/:id/:new
	var department Department
	dataId := c.Param("id")
	newName := c.Param("new")

	result := db.Model(&department).Where("id = ?", dataId).Update("Department_Name", newName)
	if result.Error != nil {
		log.Println(result.Error)
		c.String(http.StatusInternalServerError, "UPDATE error")
		c.Abort()
		return
	}

	c.String(http.StatusOK, "Department Update Complete")
}

/* 기존의 Department 삭제(D) */
func DeleteDepartment(c *gin.Context) {
	department_id := c.Param("id")

	var department Department

	// Find Department
	db.Where("id=?", department_id).Find(&department)
	if department.ID == 0 { // 테이블에 이름이 일치하는 Department가 없으면 ID = 0 으로 반환
		log.Println("Department id incorrect")

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

	result := db.Where("Department_Name = ?", name).Find(&departments)
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

/* 새로운 Employee 추가(C) */
func AddEmployee(c *gin.Context) {
	employee_name := c.Param("name")

	result := db.Create(&Employee{Employee_Name: employee_name})
	if result.Error != nil {
		log.Println(result.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error creating employee",
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": employee_name,
	})

}

/* Employee Table 불러오기(R)_By Paging */
func ReadEmployee(c *gin.Context) {
	var employees []Employee
	limit, page, sort := Paging(c)
	offset := (page - 1) * limit

	//result := db.Find(&employees)
	result := db.Limit(limit).Offset(offset).Order(sort).Find(&employees)

	if result.Error != nil {
		log.Println(result.Error)

		c.String(http.StatusInternalServerError, "READ error")
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, employees)
}

/* 기존의 Employee 내용 수정(U) */
func UpdateEmployee(c *gin.Context) {
	var employee Employee
	dataId := c.Param("id")
	newName := c.Param("new")

	result := db.Model(&employee).Where("id = ?", dataId).Update("Employee_Name", newName)
	if result.Error != nil {
		log.Println(result.Error)
		c.String(http.StatusInternalServerError, "UPDATE error")
		c.Abort()
		return
	}

	c.String(http.StatusOK, "Employee Update Complete")
}

/* 기존의 Emplpyee 삭제(D) */
func DeleteEmployee(c *gin.Context) {
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

	db.Delete(&employee)
	c.JSON(http.StatusOK, gin.H{
		"msg": "Delete Complete",
	})
}

/* n일 이내 입사한 사원 조회 */
func SearchEmployeeByDay(c *gin.Context) {
	n := c.Param("days")
	var employees []Employee

	result := db.Where("TO_DAYS(SYSDATE()) - TO_DAYS(created_at) <= ?", n).Find(&employees)
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

	result := db.Where("Employee_Name = ?", name).Find(&employees)
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
		"msg":             "employee exited by department",
		"department_name": department.Department_Name,
	})
}

/* 부서 내 소속된 사원 목록 출력 */
func ReadEmployeeInDepartment(c *gin.Context) {
	did := c.Param("did")

	var employees []Employee
	var department Department

	result := db.Where("id=?", did).Find(&department)

	limit, page, sort := Paging(c)
	offset := (page - 1) * limit

	if result.Error != nil {
		log.Println(result.Error.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": result.Error.Error(),
		})
		c.Abort()
		return
	}
	log.Println(department)

	//db.Model(&department).Association("Department_Employees").Find(&employees)
	db.Limit(limit).Offset(offset).Order(sort).Model(&department).Association("Department_Employees").Find(&employees)

	c.JSON(http.StatusOK, gin.H{
		"department id":   department.ID,
		"department name": department.Department_Name,
	})
	c.JSON(http.StatusOK, employees)
}

/* 페이징 처리 부분. HTTP Request에 Query를 통해서 변수를 받아온다 */
func Paging(c *gin.Context) (limit int, page int, sort string) { // return은 limit, page, sort
	sort = "id asc" // id는 모든 table에 있다는 점 이용해서 일단 id로 설정
	query := c.Request.URL.Query()

	for key, value := range query {
		fmt.Println(key, value)
		queryValue := value[len(value)-1]
		switch key {
		case "limit":
			limit, _ = strconv.Atoi(queryValue)
			break
		case "page":
			page, _ = strconv.Atoi(queryValue)
			break
			//case "sort":
		}
	}

	return
}

/* 계정 정보를 반환하는 Profile Handler 작성(미들웨어에서 검증이 끝나면 이를 반환할 수 있도록 함. 검증용) */
func Profile(c *gin.Context) {
	var account Account

	account_id, _ := c.Get("account_Id")
	result := db.Where("Account_Id=?", account_id.(string)).First(&account)

	if result.Error == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "account not found",
		})
		c.Abort()
		return
	}

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "could not get account profile",
		})
		c.Abort()
		return
	}

	account.Account_Pwd = "temp pwd value"
	c.JSON(http.StatusOK, account)

}

/* 토큰을 사용한 미들웨어에서의 계정 검증 */
func AuthorizeAccount() gin.HandlerFunc {
	return func(c *gin.Context) { // Handler를 return
		clientToken := c.Request.Header.Get("Authorization") // Context의 header 내용 중 key가 "Authorization"인 내용의 value를 가져옴 --> 이게 Token이 됨!
		if clientToken == "" {                               // No Header
			c.JSON(http.StatusForbidden, gin.H{
				"msg": "No Authorization header provided",
			})
			c.Abort()
			return
		}

		// Bearer Authentication 방식을 사용하기로 함
		extractedToken := strings.Split(clientToken, "Bearer ")
		if len(extractedToken) == 2 { // {"Bearer ", [토큰 내용 string]}
			clientToken = strings.TrimSpace(extractedToken[1])
		} else { // Invalid Token Format
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "Incorrect Format of Authorization Token",
			})
			c.Abort()
			return
		}

		claims, err := JWT.ValidateToken(clientToken)
		if err != nil { // Invalid Token
			c.JSON(http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}

		c.Set("account_Id", claims.Account_Id)

		c.Next()
	}
}
