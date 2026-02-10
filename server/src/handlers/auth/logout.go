package auth

import (
	"IAM-server/src/utils/env"
	"net/http"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	// deleting cookie with the refresh token
	cookie := http.Cookie{
		Name:     env.REFRESH_COOKIE_NAME,
		Value:    "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully logged out!"))
}
