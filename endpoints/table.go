package endpoints

import (
	"net/http"

	"github.com/jesses-code-adventures/every_log/db"
)

type TableHandler struct {
	Db *db.Db
}

func (t TableHandler) ServeJson(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		resp, err := t.Db.ShowTables();
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
