package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var db *gorm.DB
var err error

func main() {
	err := InitDB()
	if err != nil {
		log.Println(err.Error())
		panic("DB init error")
	}
	err = godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := SetupRouter()
	//r.RunTLS(fmt.Sprintf(":%s", os.Getenv("PORT")), "server.crt", "server.key")

	r.Run(fmt.Sprintf(":%s", os.Getenv("PORT")))
}
