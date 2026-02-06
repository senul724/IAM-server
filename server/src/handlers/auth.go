package handlers

import (
	"IAM-server/src/types"
	"encoding/json"
	"net/http"
)

func Login(w http.ResponseWriter, r *http.Request) {
	var info types.UserCredential

	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

// func Register(w http.ResponseWriter, r *http.Request) {
// 	var info types.UserCredential
//
// 	err := json.NewDecoder(r.Body).Decode(&info)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
//
// 	hashedPwd, err := utils.HashPassword(info.PWD)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
//
// 	qry := "INSERT INTO user() VALUES($1,$2)"
//
// 	connections.DBCon.DB.Exec()
//
// }
