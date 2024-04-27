package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jesses-code-adventures/every_log/db"
)

type ProjectHandler struct {
	Db *db.Db
}

func (p ProjectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ProjectHandler")
	accept := r.Header.Get("Accept")
	switch accept {
	case "application/json":
		p.ServeJson(w, r)
		return
	default:
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
}

func (p ProjectHandler) ServeJson(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		id, err := p.createProject(r)
		if err != nil {
			if err.Error() == "Db error" || err.Error() == "failed to convert to JSON" {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			} else if err.Error() == "Project Already Exists" {
				http.Error(w, err.Error(), http.StatusConflict)
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(err.Error()))
				return
			} else {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte(err.Error()))
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (p ProjectHandler) createProject(r *http.Request) ([]byte, error) {
	user_id := r.Header.Get("user_id")
	if user_id == "" {
		return nil, errors.New("User id is required")
	}
	arr := make([]byte, 0)
	body := r.Body
	defer body.Close()
	var parsedBody struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	resp, err := p.Db.CreateProject(user_id, parsedBody.Name, parsedBody.Description)
	if err != nil {
		return nil, err
	}
	arr, err = json.Marshal(resp)
	return arr, err
}
