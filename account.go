package main

import (
	"net/http"
	"os"
	"strings"

	Oauth "github.com/dunebi/myapi-oauth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	Email string `json:"email"`
	CA    string
}

func (account *Account) DbProcess(c *gin.Context) {
	var dbAccount Account
	db.Where("Email = ? AND CA = ?", account.Email, account.CA).Find(&dbAccount)

	// DB계정이 없다면 Register
	if dbAccount.ID == 0 {
		result := db.Create(&account)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"msg": "Error on Creating Account",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"msg":         "New account created. Please re-login",
			"accountInfo": account,
		})
		return
	}

	// DB에 계정이 있다면 Login(JWT토큰 발행)
	jwtToken, err := GenerateToken(account.Email, account.CA)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"msg": "Error on Create JWT token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"JWT":     jwtToken,
		"account": dbAccount,
	})
}

func Login() gin.HandlerFunc {
	oauthFunc := Oauth.LoginProcess()

	return func(c *gin.Context) {
		newCA := c.Param("CA")
		newCA = strings.ToUpper(newCA)

		url, err := oauthFunc(newCA)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"msg": err.Error(),
			})
			return
		}
		c.Redirect(http.StatusPermanentRedirect, url)
	}
}

func LoginCallback() gin.HandlerFunc {
	oauthFunc := Oauth.LoginCallbackProcess(&Account{})

	return func(c *gin.Context) {
		code := c.Query("code")
		accountResp, err := oauthFunc(code)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"msg": err.Error(),
			})
			return
		}

		account := accountResp.Interface().(*Account)
		account.CA = os.Getenv("CA")

		account.DbProcess(c)
	}
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

		claims, err := ValidateToken(clientToken)
		if err != nil { // Invalid Token
			c.JSON(http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}

		c.Set("email", claims.Email)

		c.Next()
	}
}
