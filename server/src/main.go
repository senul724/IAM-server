package main

import (
	"IAM-server/src/connections"
	"IAM-server/src/handlers"
	"IAM-server/src/handlers/auth"
	"IAM-server/src/handlers/tokens"
	"IAM-server/src/utils/env"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	//loading env
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	env.LoadEnv()

	// etablishing DB connection
	db, err := connections.ConnectDB()
	if err != nil {
		panic(err)
	}

	log.Println("DB connected successfuly")
	connections.DBCon.SetDB(db)

	// REST multiplexer
	router := http.NewServeMux()

	//handlers
	router.HandleFunc("POST /user/login", auth.Login)
	router.HandleFunc("POST /user/register", auth.RegisterUser)
	router.HandleFunc("POST /register", handlers.RegisterSite)

	router.HandleFunc("POST /token/refresh", tokens.IssueAccess)

	log.Printf("Server starting on %s", env.PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", env.PORT), router))
}
