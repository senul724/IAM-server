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

type QueryData struct {
	UserId    string `json:"user_id"`
	Name      string `json:"name"`
	PhotoUrl  string `json:"photo_url"`
	HashedPwd string `json:"hashed_password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var credentials types.UserCredential
	var queryData QueryData

	db := connections.DBCon.DB

	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	qry := `SELECT s.hashed_password, u.id, u.name, u.photo_url
	FROM user_site s 
	INNER JOIN public.user u
	ON s.user_id = u.id 
	WHERE u.email = $1 AND s.site_domain = $2`

	// creating retainer for nullable field
	var photoRet sql.NullString
	var nameRet sql.NullString

	scanErr := db.QueryRow(qry, credentials.Email, credentials.Site).Scan(
		&queryData.HashedPwd,
		&queryData.UserId,
		&nameRet,
		&photoRet,
	)

	// setting retained values and other queryData values
	queryData.PhotoUrl = utils.HadleNullSqlString(&photoRet)
	queryData.Name = utils.HadleNullSqlString(&nameRet)

	// Scanning for the user in the DB
	if scanErr == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if scanErr != nil {
		http.Error(w, scanErr.Error(), http.StatusBadRequest)
		return
	}

	// validation credentials
	pwdMatchErr := bcrypt.CompareHashAndPassword([]byte(queryData.HashedPwd), []byte(credentials.PWD))
	if pwdMatchErr != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// creating jwt
	userdata := types.UserData{
		Name:     queryData.Name,
		PhotoUrl: queryData.PhotoUrl,
		Email:    credentials.Email,
	}
	token, jwtError := utils.CreateRefreshToken(queryData.UserId, &userdata)
	if jwtError != nil {
		http.Error(w, "Failed to generate token:"+jwtError.Error(), http.StatusInternalServerError)
		return
	}

	// setting cookies
	cookie := http.Cookie{
		Name:     "iam-refresh",
		Value:    token,
		MaxAge:   5000,
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully logged in!"))
}
