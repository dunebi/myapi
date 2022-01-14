package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

/* DB를 생성(test를 위함) */
func InitDB() (err error) {
	dsn := "root:1234@tcp(127.0.0.1:3306)/myapi?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return
	}

	return
}

/* Table 생성 */
func InitTable(c *gin.Context) {
	err := db.AutoMigrate(&Account{}, &Department{}, &Employee{}) // DB Table 생성
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Cannot Init Table",
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "Table Init",
	})
}

/* Table 전체삭제 */
func DeleteTable(c *gin.Context) {
	err := db.Migrator().DropTable(&Account{}, &Department{}, &Employee{}, "employee_departments") // DB Table 삭제
	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Cannot Delete Table",
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "Table Delete",
	})
}

/* 페이징 처리 부분. HTTP Request에 Query를 통해서 변수를 받아온다 */
func Paging(c *gin.Context) (limit int, page int, sort string) { // return은 limit, page, sort
	sort = "id asc" // id는 모든 table에 있다는 점 이용해서 일단 id로 설정 후 오름차순으로
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
