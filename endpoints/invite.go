package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jesses-code-adventures/every_log/db"
	"github.com/jesses-code-adventures/every_log/error_msgs"
)

type ProjectInviteHandler struct {
	Db     *db.Db
	Logger *log.Logger
}

func (p ProjectInviteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (p ProjectInviteHandler) ServeJson(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		id, err := p.create(r)
		if err != nil {
			status := error_msgs.GetErrorHttpStatus(err)
			http.Error(w, error_msgs.JsonifyError(err.Error()), status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"id": %s}`, id)))
	case http.MethodGet:
		projectInvites, err := p.get(r)
		if err != nil {
			status := error_msgs.GetErrorHttpStatus(err)
			http.Error(w, error_msgs.JsonifyError(err.Error()), status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(projectInvites)
	default:
		http.Error(w, error_msgs.JsonifyError(error_msgs.UNACCEPTABLE_HTTP_METHOD), http.StatusMethodNotAllowed)
	}
}

func (p ProjectInviteHandler) create(r *http.Request) ([]byte, error) {
	userId := r.Header.Get("user_id")
	if userId == "" {
		return nil, errors.New(error_msgs.USER_ID_REQUIRED)
	}
	apiKey := r.Header.Get("api_key")
	if apiKey == "" {
		return nil, errors.New(error_msgs.API_KEY_REQUIRED)
	}
	projectId := r.PathValue("project_id")
	if projectId == "" {
		return nil, errors.New(error_msgs.GetRequiredMessage("project_id"))
	}
	arr := make([]byte, 0)
	body := r.Body
	defer body.Close()
	var parsedBody struct {
		ToUserId  string `json:"to_user_id"`
	}
	err := json.NewDecoder(body).Decode(&parsedBody)
	if err != nil {
		p.Logger.Println(err)
		return nil, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	resp, err := p.Db.CreateProjectInvite(userId, parsedBody.ToUserId, projectId, apiKey)
	if err != nil {
		return nil, err
	}
	arr, err = json.Marshal(resp)
	if err != nil {
		p.Logger.Println(err)
		return nil, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	return arr, err
}

func (p ProjectInviteHandler) get(r *http.Request) ([]byte, error) {
	userId := r.Header.Get("user_id")
	if userId == "" {
		return nil, errors.New(error_msgs.USER_ID_REQUIRED)
	}
	arr := make([]byte, 0)
	body := r.Body
	defer body.Close()
	var parsedBody struct {
		fromUserId *string
		toUserId   *string
		projectId  *string
		status     *string
		from       *time.Time
		to         *time.Time
	}
	err := json.NewDecoder(body).Decode(&parsedBody)
	if err != nil {
		p.Logger.Println(err)
		return nil, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	resp, err := p.Db.GetProjectInvites(userId, parsedBody.fromUserId, parsedBody.toUserId, parsedBody.projectId, parsedBody.status, parsedBody.from, parsedBody.to)
	if err != nil {
		return nil, err
	}
	arr, err = json.Marshal(resp)
	if err != nil {
		p.Logger.Println(err)
		return nil, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	return arr, err
}
