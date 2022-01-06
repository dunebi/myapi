package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
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

type JwtClaim struct {
	Account_Id string
	jwt.StandardClaims
}

/* JWT는 Header, Payload, Signature로 이루어짐 */
/* Payload는 "전송하는 내용" 정도의 뜻을 담고 있음. 로그인 정보라고 볼 수 있음 */
type LoginPayload struct {
	Account_Id  string `json:"account_id"`
	Account_Pwd string `json:"account_pwd"`
}

type LoginResponse struct {
	Token string `json:"token"` // 로그인 시 응답으로 받는 Token, JWT Token이 될 것으로 보임
}

var db *gorm.DB
var err error

func main() {
	dsn := "root:1234@tcp(127.0.0.1:3306)/myapi?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect db")
	}

	//db.AutoMigrate(&Account{}, &Department{}, &Employee{}) // DB Table 생성
	//db.Migrator().DropTable(&Account{}, &Department{}, &Employee{}, "employee_departments") // DB Table 삭제
	//db.Create(&Employee{Employee_Name: "KMC"})
	//db.Create(&Employee{Employee_Name: "IJY"})
	//db.Create(&Employee{Employee_Name: "PMS"})
	//db.Create(&Department{Department_Name: "A"})
	//db.Create(&Department{Department_Name: "B"})

	//var employee Employee
	var employees []Employee
	//var department Department

	db.Find(&employees)

	fmt.Println(employees)

	//db.Where("Department_Name=?", "B").Find(&department)
	//db.Model(&department).Association("Department_Employees").Find(&employees)
	//fmt.Println(employees)

	// 기존에 존재하는 employee, department로 association 만들기
	// 이렇게 할 경우 employee_departments Table에 새로 Association이 추가됨!
	//db.Where("Employee_Name=?", "KMC").Find(&employee)
	//db.Where("Department_Name=?", "B").Find(&department)
	//fmt.Println(department)
	//db.Model(&employee).Association("Employee_Departments").Append(&department)

	//db.Where("Employee_Name=?", "PMS").Find(&employee2)
	//db.Model(&employee2).Association("Employee_Departments").Append(&department)

	//db.Model(&employee).Association("Employee_Departments").Append(&department)
	// department가 없어도 association이 추가됨. 나중에 처리해야 할 부분????

	//db.Model(&employees).Association("Employee_Departments").Find(&department)
	//fmt.Println(employees)

	//db.Where("Department_Name=?", "B").Delete(&Department{})
	//db.Select("employee_departments").Delete(&department)
	//db.Model(&employee).Association("Employee_Departments").Find(&department)
	//db.Model(&employee).Association("Employee_Departments").Delete(&department) // 종속성 삭제
	//db.Model(&employee2).Association("Employee_Departments").Clear()

	// API Server Open
	r := SetupRouter()

	r.Run(":8080")
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

/* 로그인 후 사용할 JWT 토큰을 생성함 */
func GenerateToken(account_id string) (signedToken string, err error) {
	claims := &JwtClaim{ // Account ID와 만료에 대한 정보를 담고 있음
		Account_Id: account_id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(15 * time.Minute).Unix(), // 15분 뒤 만료
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)   // token 생성
	signedToken, err = token.SignedString([]byte("SecreteCode")) // Secretecode는 토큰 자체 인증키(자체 비밀번호 느낌)

	if err != nil {
		return
	}
	return
}

/* JWT 토큰 검증 */
func ValidateToken(signedToken string) (claims *JwtClaim, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JwtClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte("SecreteCode"), nil
		},
	)

	if err != nil {
		return
	}

	claims, ok := token.Claims.(*JwtClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("jwt is expired")
		return
	}
	return
}

/* API 세팅 */
func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) { // GET(code, handler)
		c.String(200, "pong")
	})

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
			department.PUT("/:id/:new", UpdateDepartment)
			department.POST("/:department_name", AddDepartment)
			department.DELETE("/:department_name", DeleteDepartment)
		}
		employee := api.Group("/employee").Use(AuthorizeAccount())
		{
			employee.GET("/", ReadEmployee)
			employee.PUT("/:id/:new", UpdateEmployee)
			employee.POST("/:employee_name", AddEmployee)
			employee.DELETE("/:employee_name", DeleteEmployee)
		}

		/*
			protected.GET("/profile", Profile)
			protected.POST("/addDepartment/:department_name", AddDepartment)
			protected.POST("/deleteDepartment/:department_name", DeleteDepartment)
			protected.POST("/addEmployee/:employee_name", AddEmployee)
			protected.POST("/deleteEmployee/:employee_name", DeleteEmployee)
			protected.POST("/addEmployeeDepartment/:employee_name/:department_name", AddEmployeeDepartment)
			protected.POST("/deleteEmployeeDepartment/:employee_name/:department_name", DeleteEmployeeDepartment)
		*/
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
	var payload LoginPayload
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
	signedToken, err := GenerateToken(account.Account_Id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error signing token",
		})
		c.Abort()
		return
	}

	tokenResponse := LoginResponse{ // string인 LoginResponse를 JSON으로 반환하기 위함
		Token: signedToken,
	}
	fmt.Println(tokenResponse) // 형태 확인용
	c.JSON(http.StatusOK, tokenResponse)
}

/* 토큰 인증된 사용자(일명 관리자)가 사용할 API Handler들(CRUD) */

/* 새로운 Department를 추가(C) */
func AddDepartment(c *gin.Context) {
	department_name := c.Param("department_name")

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
func ReadDepartment(c *gin.Context) { // localhost:8080/api/department/ (GET)
	var departments []Department
	result := db.Find(&departments)

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
	department_name := c.Param("department_name")

	var department Department

	// Find Department
	db.Where("Department_Name=?", department_name).Find(&department)
	if department.ID == 0 { // 테이블에 이름이 일치하는 Department가 없으면 ID = 0 으로 반환
		log.Println("Department name incorrect")

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error deleting department",
		})
		c.Abort()
		return
	}

	db.Delete(&department)
	c.JSON(http.StatusOK, gin.H{
		"msg": "Delete Complete",
	})

}

/* 새로운 Employee 추가(C) */
func AddEmployee(c *gin.Context) {
	employee_name := c.Param("employee_name")

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

/* Employee Table 불러오기(R) */
func ReadEmployee(c *gin.Context) {
	var employees []Employee
	result := db.Find(&employees)

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
	employee_name := c.Param("employee_name")

	var employee Employee

	// Find employee
	db.Where("Employee_Name=?", employee_name).Find(&employee)
	if employee.ID == 0 {
		log.Println("Employee name incorrect")

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

/* 사원에게 부서 만들어주기 */
func AddEmployeeDepartment(c *gin.Context) {
	employee_name := c.Param("employee_name")
	department_name := c.Param("department_name")

	var employee Employee
	var department Department

	db.Where("Employee_Name = ?", employee_name).Find(&employee)
	db.Where("Department_Name = ?", department_name).Find(&department)

	if (employee.ID == 0) || (department.ID == 0) {
		log.Println("Name error at employee or department")

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error allocate department to employee",
		})
		c.Abort()
		return
	}

	db.Model(&employee).Association("Employee_Departments").Append(&department)
	c.JSON(http.StatusOK, gin.H{
		"employee":   employee_name,
		"department": department_name,
	})
}

/* 사원을 부서에서 제외시키기 */
func DeleteEmployeeDepartment(c *gin.Context) {
	employee_name := c.Param("employee_name")
	department_name := c.Param("department_name")

	var employee Employee
	var department Department

	db.Where("Employee_Name = ?", employee_name).Find(&employee)
	db.Where("Department_Name = ?", department_name).Find(&department)

	if (employee.ID == 0) || (department.ID == 0) {
		log.Println("Name error at employee or department")

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error name at department or employee",
		})
		c.Abort()
		return
	}

	db.Model(&employee).Association("Employee_Departments").Delete(&department)
	c.JSON(http.StatusOK, gin.H{
		"msg":             "employee exited by department",
		"department_name": department_name,
	})
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
		if clientToken == "" {
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
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "Incorrect Format of Authorization Token",
			})
			c.Abort()
			return
		}

		claims, err := ValidateToken(clientToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}

		c.Set("account_Id", claims.Account_Id)

		//fmt.Println("JWT Validation Complete")
		c.Next()
	}
}
