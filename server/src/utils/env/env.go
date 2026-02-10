package env

import "os"

var (
	PORT                string
	DB_URI              string
	REFRESH_KEY         string
	ACCESS_KEY          string
	REFRESH_COOKIE_NAME string
)

func LoadEnv() {
	DB_URI = os.Getenv("DB_URI")
	REFRESH_KEY = os.Getenv("RERESH_KEY")
	ACCESS_KEY = os.Getenv("ACCESS_KEY")
	REFRESH_COOKIE_NAME = os.Getenv("REFRESH_COOKIE_NAME")
	PORT = os.Getenv("PORT")
}
