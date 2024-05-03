package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jesses-code-adventures/every_log/db"
	"github.com/jesses-code-adventures/every_log/error_msgs"
)

type LogHandler struct {
	Db *db.Db
}

func (p LogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (p LogHandler) ServeJson(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		id, err := p.createLog(r)
		if err != nil {
			fmt.Printf(fmt.Sprintf("got error: %s", err))
			status := error_msgs.GetErrorHttpStatus(err)
			http.Error(w, error_msgs.JsonifyError(err.Error()), status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"id": %s}`, id)))
	case http.MethodGet:
		logs, err := p.getLogs(r)
		if err != nil {
			status := error_msgs.GetErrorHttpStatus(err)
			http.Error(w, error_msgs.JsonifyError(err.Error()), status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(logs)
	default:
		http.Error(w, error_msgs.JsonifyError(error_msgs.UNACCEPTABLE_HTTP_METHOD), http.StatusMethodNotAllowed)
	}
}

func (p LogHandler) createLog(r *http.Request) ([]byte, error) {
	userId := r.Header.Get("user_id")
	if userId == "" {
		return nil, errors.New(error_msgs.USER_ID_REQUIRED)
	}
	apiKey := r.Header.Get("api_key")
	if apiKey == "" {
		return nil, errors.New(error_msgs.API_KEY_REQUIRED)
	}
	arr := make([]byte, 0)
	body := r.Body
	defer body.Close()
	var parsedBody struct {
		ProjectId string  `json:"project_id"`
		LevelId   int     `json:"level_id"`
		ProcessId *string `json:"process_id"`
		Message   string  `json:"message"`
		Traceback *string `json:"traceback"`
	}
	err := json.NewDecoder(body).Decode(&parsedBody)
	if err != nil {
		fmt.Println(err) //TODO: Use a logger
		return nil, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	resp, err := p.Db.CreateLog(userId, parsedBody.ProjectId, parsedBody.LevelId, parsedBody.ProcessId, parsedBody.Message, parsedBody.Traceback, apiKey)
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

func (p LogHandler) getLogs(r *http.Request) ([]byte, error) {
	userId := r.Header.Get("user_id")
	if userId == "" {
		return nil, errors.New(error_msgs.USER_ID_REQUIRED)
	}
	arr := make([]byte, 0)
	body := r.Body
	defer body.Close()
	var parsedBody struct {
		ProjectId *string    `json:"project_id"`
		LevelId   *int       `json:"level_id"`
		ProcessId *string    `json:"process_id"`
		OrgId     *string    `json:"org_id"`
		Message   *string    `json:"message"`
		Traceback *string    `json:"traceback"`
		From      *time.Time `json:"from"`
		To        *time.Time `json:"to"`
	}
	err := json.NewDecoder(body).Decode(&parsedBody)
	if err != nil {
		fmt.Println(err) //TODO: Use a logger
		return nil, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	resp, err := p.Db.GetLogs(userId, parsedBody.ProjectId, parsedBody.LevelId, parsedBody.ProcessId, parsedBody.OrgId, parsedBody.From, parsedBody.To)
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
