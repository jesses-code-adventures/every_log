package endpoints

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jesses-code-adventures/every_log/db"
)

type UserHandler struct {
	Db *db.Db
}

// Creates a new user in the db and returns their id in a JSON response
func (u UserHandler) CreateUser(r *http.Request) ([]byte, error) {
	arr := make([]byte, 0)
	body := r.Body
	defer body.Close()
	// Extract fields from JSON body
	var request struct {
		Email     string  `json:"email"`
		FirstName string  `json:"first_name"`
		LastName  *string `json:"last_name"`
		Password  string  `json:"password"`
	}
	err := json.NewDecoder(body).Decode(&request)
	if err != nil {
		return arr, err
	}
	id, err := u.Db.CreateUser(request.Email, request.FirstName, request.LastName, request.Password)
	if err != nil {
		if err.Error() == "Email already exists" {
			return arr, errors.New("Email already exists")
		}
		return arr, errors.New("Db error")
	}
	response := struct {
		ID string `json:"id"`
	}{
		ID: id,
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return nil, errors.New("failed to convert response to json")
	}
	return jsonBytes, nil
}

func (u UserHandler) ServeJson(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		id, err := u.CreateUser(r)
		if err != nil {
			if err.Error() == "Db error" || err.Error() == "failed to convert to JSON" {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else if err.Error() == "Email already exists" {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			} else {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
