package auth

import (
	"IAM-server/src/utils"
	"net/http"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	// deleting cookie with the refresh token
	cookie := http.Cookie{
		Name:     utils.RefreshCookieName,
		Value:    "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully logged out!"))
}
