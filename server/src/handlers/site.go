package handlers

import (
	"IAM-server/src/connections"
	"IAM-server/src/types"
	"encoding/json"
	"net/http"
)

func RegisterSite(w http.ResponseWriter, r *http.Request) {
	var info types.SiteData
	db := connections.DBCon.DB

	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	qry := "INSERT INTO site VALUES($1,$2,$3,$4)"

	_, exec_err := db.Exec(qry, info.Domain, info.Name, info.Description, info.PhotoUrl)

	if exec_err != nil {
		http.Error(w, exec_err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Site added successfully"))
}
