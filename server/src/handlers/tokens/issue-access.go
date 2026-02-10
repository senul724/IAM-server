package tokens

import (
	"IAM-server/src/utils"
	"net/http"
)

func IssueAccess(w http.ResponseWriter, r *http.Request) {
	// check for refresh token
	cookieErr := utils.CheckRefreshToken(r)
	if cookieErr != nil {
		http.Error(w, cookieErr.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("access token"))

}
