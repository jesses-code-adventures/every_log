package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jesses-code-adventures/every_log/db"
	"github.com/jesses-code-adventures/every_log/error_msgs"
)

type ApiKeyHandler struct {
	Db *db.Db
}

func (p ApiKeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (p ApiKeyHandler) ServeJson(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		projectId := r.PathValue("project_id")
		if projectId == "" {
			http.Error(w, error_msgs.JsonifyError(error_msgs.GetRequiredMessage("project_id")), http.StatusBadRequest)
			return
		}
		key, err := p.createApiKey(r, projectId)
		if err != nil {
			fmt.Printf(fmt.Sprintf("got error: %s", err))
			status := error_msgs.GetErrorHttpStatus(err)
			http.Error(w, error_msgs.JsonifyError(err.Error()), status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"key": %s}`, key)))
	default:
		http.Error(w, error_msgs.JsonifyError(error_msgs.UNACCEPTABLE_HTTP_METHOD), http.StatusMethodNotAllowed)
	}
}

func (p ApiKeyHandler) createApiKey(r *http.Request, projectId string) ([]byte, error) {
	user_id := r.Header.Get("user_id")
	if user_id == "" {
		return nil, errors.New(error_msgs.USER_ID_REQUIRED)
	}
	arr := make([]byte, 0)
	resp, err := p.Db.CreateApiKey(user_id, projectId)
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
