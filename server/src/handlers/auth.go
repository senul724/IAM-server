package handlers

import (
	"IAM-server/src/connections"
	"IAM-server/src/types"
	"IAM-server/src/utils"
	"database/sql"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type payload struct {
	UserId    string `json:"user_id"`
	Name      string `json:"name"`
	PhotoUrl  string `json:"photo_url"`
	HashedPwd string `json:"hashed_password"`
	LastLogin int    `json:"last_login"`
	Token     string `json:"token"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var info types.UserCredential
	var payload payload

	db := connections.DBCon.DB

	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	qry := `SELECT s.hashed_password, u.id, s.last_login, u.name, u.photo_url
	FROM user_site s 
	INNER JOIN public.user u
	ON s.user_id = u.id 
	WHERE u.email = $1 AND s.site_domain = $2`

	var photo sql.NullString

	scanErr := db.QueryRow(qry, info.Email, info.Site).Scan(
		&payload.HashedPwd,
		&payload.UserId,
		&payload.LastLogin,
		&payload.Name,
		&photo,
	)

	if scanErr == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if scanErr != nil {
		http.Error(w, scanErr.Error(), http.StatusBadRequest)
		return
	}

	pwdMatchErr := bcrypt.CompareHashAndPassword([]byte(payload.HashedPwd), []byte(info.PWD))
	if pwdMatchErr != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	userdata := types.UserData{
		Name:     payload.Name,
		PhotoUrl: payload.PhotoUrl,
		Email:    info.Email,
	}
	token, jwtError := utils.CreateToken(payload.UserId, &userdata)
	if jwtError != nil {
		http.Error(w, "Failed to generate token:"+jwtError.Error(), http.StatusInternalServerError)
		return
	}

	payload.Token = token

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "json")
	json.NewEncoder(w).Encode(payload)
}
