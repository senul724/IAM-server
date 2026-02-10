package auth

import (
	"IAM-server/src/connections"
	"IAM-server/src/utils"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type registerData struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Pwd      string `json:"pwd"`
	Site     string `json:"site"`
	PhotoUrl string `json:"photo_url"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	// checking if already logged in
	rTokenError := utils.CheckRefreshToken(r)

	// reverting if there is no error present which means a valid refresh token is present
	if rTokenError == nil {
		http.Error(w, "Already logged in", http.StatusBadRequest)
		return
	}

	var data registerData

	userId := uuid.New()
	db := connections.DBCon.DB

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pwdHash, hashError := bcrypt.GenerateFromPassword([]byte(data.Pwd), bcrypt.DefaultCost)
	if hashError != nil {
		http.Error(w, hashError.Error(), http.StatusBadRequest)
		return
	}

	qry := `INSERT INTO public.user VALUES($1, $2, $3, $4)`

	_, insertErr := db.Exec(qry, userId, data.Email, data.Name, data.PhotoUrl)
	if insertErr != nil {
		http.Error(w, insertErr.Error(), http.StatusBadRequest)
		return
	}

	qry = `INSERT INTO user_site VALUES($1, $2, $3, $4, $5)`

	_, siteInsertError := db.Exec(qry, userId, data.Site, pwdHash, time.Now().Unix(), "dasdas")
	if siteInsertError != nil {
		http.Error(w, siteInsertError.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User added succesfully!"))
}
