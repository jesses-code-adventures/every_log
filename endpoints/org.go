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

type OrgHandler struct {
	Db     *db.Db
	Logger *log.Logger
}

func (o OrgHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	switch accept {
	case "application/json":
		o.ServeJson(w, r)
		return
	default:
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
}

func (o OrgHandler) ServeJson(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		id, err := o.create(r)
		if err != nil {
			status := error_msgs.GetErrorHttpStatus(err)
			http.Error(w, error_msgs.JsonifyError(err.Error()), status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"id": %s}`, id)))
	case http.MethodGet:
		logs, err := o.get(r)
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

func (o OrgHandler) create(r *http.Request) ([]byte, error) {
	userId := r.Header.Get("user_id")
	if userId == "" {
		return nil, errors.New(error_msgs.USER_ID_REQUIRED)
	}
	arr := make([]byte, 0)
	body := r.Body
	defer body.Close()
	var parsedBody struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		LocationId  *string `json:"location_id"`
	}
	err := json.NewDecoder(body).Decode(&parsedBody)
	if err != nil {
		o.Logger.Println(err)
		return nil, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	resp, err := o.Db.CreateOrg(userId, parsedBody.Name, parsedBody.Description, parsedBody.LocationId)
	if err != nil {
		return nil, err
	}
	arr, err = json.Marshal(resp)
	if err != nil {
		o.Logger.Println(err)
		return nil, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	return arr, err
}

func (o OrgHandler) get(r *http.Request) ([]byte, error) {
	userId := r.Header.Get("user_id")
	if userId == "" {
		return nil, errors.New(error_msgs.USER_ID_REQUIRED)
	}
	arr := make([]byte, 0)
	body := r.Body
	defer body.Close()
	var parsedBody struct {
		OrgId *string    `json:"org_id"`
		Name  *string    `json:"name"`
		From  *time.Time `json:"from"`
		To    *time.Time `json:"to"`
	}
	err := json.NewDecoder(body).Decode(&parsedBody)
	if err != nil {
		o.Logger.Println(err)
		return nil, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	resp, err := o.Db.GetOrgs(userId, parsedBody.OrgId, parsedBody.Name, parsedBody.From, parsedBody.To)
	if err != nil {
		return nil, err
	}
	arr, err = json.Marshal(resp)
	if err != nil {
		o.Logger.Println(err)
		return nil, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	return arr, err
}
