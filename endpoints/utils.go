package endpoints

import (
	"errors"
	"net/http"

	"github.com/jesses-code-adventures/every_log/error_msgs"
)

type EndpointHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	ServeJson(w http.ResponseWriter, r *http.Request)
}

func BasicValidateRequest(w http.ResponseWriter, r *http.Request) error {
	accept := r.Header.Get("Accept")
	if accept != "application/json" {
		http.Error(w, error_msgs.JsonifyError("Invalid Accept Header"), http.StatusBadRequest)
		return errors.New("Accept")
	}
	apiKey := r.Header.Get("x-api-key")
	authorization := r.Header.Get("Authorization")
	if !(r.Method == http.MethodPost) && !(r.RequestURI == "/user") && !(r.RequestURI == "/authenticate") && apiKey == "" && authorization == "" {
		http.Error(w, error_msgs.JsonifyError("Missing authorization header and api key"), http.StatusUnauthorized)
		return errors.New("Auth")
	}
	return nil
}
