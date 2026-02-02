package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var users = make(map[int]User)

func main() {
	mux := http.NewServeMux()

	//handlers
	mux.HandleFunc("/", handler)
	mux.HandleFunc("POST /add", addUser)

	log.Println("Server starting on :8000")
	log.Fatal(http.ListenAndServe(":8000", mux))
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the custom IAM server")
	log.Println("GET at '/'")
}

func addUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	users[user.ID] = user

	w.WriteHeader(http.StatusAccepted)
}
