package endpoints

import (
	"net/http"

	"github.com/jesses-code-adventures/every_log/db"
)

type CheckHandler struct {
	Db *db.Db
}

func (t CheckHandler) ServeJson(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tableName := r.URL.Query().Get("table")
		if tableName == "" {
			http.Error(w, "Table name is required", http.StatusBadRequest)
			return
		}
		resp, err := t.Db.CheckTableExists(tableName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
