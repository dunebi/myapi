package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

var oauth2ConfigGoogle *oauth2.Config
var oauth2ConfigFacebook *oauth2.Config
var oauth2ConfigGithub *oauth2.Config
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

	oauth2ConfigGoogle = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	oauth2ConfigFacebook = &oauth2.Config{
		ClientID:     os.Getenv("FACEBOOK_CLIENT_ID"),
		ClientSecret: os.Getenv("FACEBOOK_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("FACEBOOK_REDIRECT_URL"),
		Scopes:       []string{"email"},
		Endpoint:     facebook.Endpoint,
	}

	oauth2ConfigGithub = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
		Scopes:       []string{"user"},
		Endpoint:     github.Endpoint,
	}

	r := SetupRouter()
	//r.RunTLS(fmt.Sprintf(":%s", os.Getenv("PORT")), "server.crt", "server.key")

	r.Run(fmt.Sprintf(":%s", os.Getenv("PORT")))
}
