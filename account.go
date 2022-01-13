/* 관리자 계정 Account */
package main

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/dunebi/myapi/JWT"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Account Table
type Account struct {
	gorm.Model
	Account_Id  string
	Account_Pwd string
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
