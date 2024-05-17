package endpoints

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/jesses-code-adventures/every_log/db"
	"github.com/jesses-code-adventures/every_log/error_msgs"
)

type UserHandler struct {
	Db     *db.Db
	Logger *log.Logger
}

func (u UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	switch accept {
	case "application/json":
		u.ServeJson(w, r)
		return
	default:
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
}

func (u UserHandler) ServeJson(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		id, err := u.createUser(r)
		if err != nil {
			status := error_msgs.GetErrorHttpStatus(err)
			http.Error(w, error_msgs.JsonifyError(err.Error()), status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(id)
	default:
		http.Error(w, error_msgs.JsonifyError(error_msgs.UNACCEPTABLE_HTTP_METHOD), http.StatusMethodNotAllowed)
	}
}

// Creates a new user in the db and returns their id in a JSON response
func (u UserHandler) createUser(r *http.Request) ([]byte, error) {
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
		u.Logger.Println(err)
		return arr, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	id, err := u.Db.CreateUser(request.Email, request.FirstName, request.LastName, request.Password)
	if err != nil {
		return arr, err
	}
	response := struct {
		ID string `json:"id"`
	}{
		ID: id,
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		u.Logger.Println(err)
		return nil, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	return jsonBytes, nil
}
