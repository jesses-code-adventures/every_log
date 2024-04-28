package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jesses-code-adventures/every_log/db"
	"github.com/jesses-code-adventures/every_log/error_msgs"
)

type ProjectHandler struct {
	Db *db.Db
}

func (p ProjectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
			fmt.Printf(fmt.Sprintf("got error: %s", err))
			status := error_msgs.GetErrorHttpStatus(err)
			http.Error(w, error_msgs.JsonifyError(err.Error()), status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"id": %s}`, id)))
	default:
		http.Error(w, error_msgs.JsonifyError(error_msgs.UNACCEPTABLE_HTTP_METHOD), http.StatusMethodNotAllowed)
	}
}

func (p ProjectHandler) createProject(r *http.Request) ([]byte, error) {
	user_id := r.Header.Get("user_id")
	if user_id == "" {
		return nil, errors.New(error_msgs.USER_ID_REQUIRED)
	}
	arr := make([]byte, 0)
	body := r.Body
	defer body.Close()
	var parsedBody struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	err := json.NewDecoder(body).Decode(&parsedBody)
	if err != nil {
		fmt.Println(err) //TODO: Use a logger
		return nil, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	resp, err := p.Db.CreateProject(user_id, parsedBody.Name, parsedBody.Description)
	if err != nil {
		return nil, err
	}
	arr, err = json.Marshal(resp)
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return nil, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	return arr, err
}
