package endpoints

import (
	"errors"
	"net/http"

	"github.com/jesses-code-adventures/every_log/db"
	"github.com/jesses-code-adventures/every_log/error_msgs"
)

type CheckHandler struct {
	Db *db.Db
}

func (t CheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	switch accept {
	case "application/json":
		t.ServeJson(w, r)
		return
	default:
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
}

func (t CheckHandler) ServeJson(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tableName := r.URL.Query().Get("table")
		if tableName == "" {
			msg := error_msgs.GetRequiredMessage("Table name")
			status := error_msgs.GetErrorHttpStatus(errors.New(msg))
			http.Error(w, error_msgs.JsonifyError(msg), status)
			return
		}
		resp, err := t.Db.CheckTableExists(tableName)
		if err != nil {
			status := error_msgs.GetErrorHttpStatus(err)
			http.Error(w, error_msgs.JsonifyError(err.Error()), status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	default:
		http.Error(w, error_msgs.JsonifyError(error_msgs.UNACCEPTABLE_HTTP_METHOD), http.StatusMethodNotAllowed)
	}
}
