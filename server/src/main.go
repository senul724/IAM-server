package main

import (
	"IAM-server/src/connections"
	"IAM-server/src/handlers"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	db, err := connections.ConnectDB()
	if err != nil {
		panic(err)
	}

	log.Println("DB connected successfuly")
	connections.DBCon.SetDB(db)

	// REST multiplexer
	mux := http.NewServeMux()

	//handlers
	mux.HandleFunc("POST /login", handlers.Login)
	mux.HandleFunc("POST /register", handlers.RegisterSite)

	log.Println("Server starting on :8000")
	log.Fatal(http.ListenAndServe(":8000", mux))
}
