/* 관리자 계정 Account */
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dunebi/myapi/JWT"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// Account Table
type Account struct {
	gorm.Model
	Email string `json:"email"`
	CA    string
}

func LoginGithub(c *gin.Context) {
	url := oauth2ConfigGithub.AuthCodeURL("state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func LoginCallbackGithub(c *gin.Context) {
	code := c.Query("code")

	tok, err := oauth2ConfigGithub.Exchange(oauth2.NoContext, code)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"msg": "Error on get token",
		})
		return
	}

	client := oauth2ConfigGithub.Client(oauth2.NoContext, tok)
	userInfoResp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":   "Error on get usrInfo",
			"error": err.Error(),
		})
	}
	defer userInfoResp.Body.Close()
	userInfo, err := ioutil.ReadAll(userInfoResp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Error on read userinfo",
		})
		c.Abort()
		return
	}

	var account, dbAccount Account
	json.Unmarshal(userInfo, &account)

	email := account.Email
	db.Where("Email = ? AND CA = ?", email, "Github").Find(&dbAccount)

	// DB 계정이 없다면 Register
	if dbAccount.ID == 0 { // No Email Info. Auto register and request re-login
		account.CA = "Github"
		result := db.Create(&account)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"msg": "Error on creating Account",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"msg":         "New account created. Please re-login",
			"accountInfo": account,
		})
		return
	}

	// DB에 계정이 있는 경우. JWT토큰을 생성하여 반환
	jwtToken, err := JWT.GenerateToken(email)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"msg": "Error on Create JWT token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"JWT": jwtToken,
	})

} // localhost:8090/login/github

// Use ngrok
func LoginFacebook(c *gin.Context) {
	url := oauth2ConfigFacebook.AuthCodeURL("state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func LoginCallbackFacebook(c *gin.Context) {
	code := c.Query("code")

	tok, err := oauth2ConfigFacebook.Exchange(oauth2.NoContext, code)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"msg": "Error on get token",
		})
		return
	}

	client := oauth2ConfigFacebook.Client(oauth2.NoContext, tok)
	userInfoResp, err := client.Get("https://graph.facebook.com/me?locale=en_US&fields=name,email")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":   "Error on get usrInfo",
			"error": err.Error(),
		})
	}
	defer userInfoResp.Body.Close()
	userInfo, err := ioutil.ReadAll(userInfoResp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Error on read userinfo",
		})
		c.Abort()
		return
	}
	var account, dbAccount Account
	json.Unmarshal(userInfo, &account)

	email := account.Email
	db.Where("Email = ? AND CA = ?", email, "Facebook").Find(&dbAccount)

	// DB 계정이 없다면 Register
	if dbAccount.ID == 0 { // No Email Info. Auto register and request re-login
		account.CA = "Facebook"
		result := db.Create(&account)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"msg": "Error on creating Account",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"msg":         "New account created. Please re-login",
			"accountInfo": account,
		})
		return
	}

	// DB에 계정이 있는 경우. JWT토큰을 생성하여 반환
	jwtToken, err := JWT.GenerateToken(email)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"msg": "Error on Create JWT token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"JWT": jwtToken,
	})

}

/* 로그인(회원가입) 구글 */
func LoginGoogle(c *gin.Context) {
	url := oauth2ConfigGoogle.AuthCodeURL("state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func LoginCallbackGoogle(c *gin.Context) {
	code := c.Query("code")

	tok, err := oauth2ConfigGoogle.Exchange(oauth2.NoContext, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Error on make token",
			"err": err,
		})
		c.Abort()
		return
	}

	client := oauth2ConfigGoogle.Client(oauth2.NoContext, tok)
	userInfoResp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":   "Error on get usrInfo",
			"error": err.Error(),
		})
	}
	defer userInfoResp.Body.Close()
	userInfo, err := ioutil.ReadAll(userInfoResp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Error on read userinfo",
		})
		c.Abort()
		return
	}

	var account, dbAccount Account
	json.Unmarshal(userInfo, &account)

	email := account.Email
	db.Where("Email = ? AND CA = ?", email, "Google").Find(&dbAccount)

	// DB 계정이 없다면 Register
	if dbAccount.ID == 0 { // No Email Info. Auto register and request re-login
		account.CA = "Google"
		result := db.Create(&account)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"msg": "Error on creating Account",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"msg":         "New account created. Please re-login",
			"accountInfo": account,
		})
		return
	}

	// DB에 계정이 있는 경우. JWT토큰을 생성하여 반환
	jwtToken, err := JWT.GenerateToken(email)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"msg": "Error on Create JWT token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"JWT": jwtToken,
	})

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
