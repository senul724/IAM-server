package tokens

import (
	"IAM-server/src/utils"
	"encoding/json"
	"net/http"
)

func IssueAccess(w http.ResponseWriter, r *http.Request) {
	// check for refresh token
	claims, cookieErr := utils.VerifyRefreshToken(r)
	if cookieErr != nil {
		http.Error(w, cookieErr.Error(), http.StatusBadRequest)
		return
	}

	// creating the access token
	token, tokenError := utils.CreateAccessToken(claims.Subject, claims.User)
	if tokenError != nil {
		http.Error(w, tokenError.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
