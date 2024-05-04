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
	if r.Method == http.MethodPost && r.RequestURI == "/user" {
		return nil
	}
	if r.Method == http.MethodPost && r.RequestURI == "/authenticate" {
		return nil
	}
	authorization, err := r.Cookie("Authorization")
	if err != nil {
		http.Error(w, error_msgs.JsonifyError("Missing authorization header"), http.StatusUnauthorized)
		return errors.New(error_msgs.UNAUTHORIZED)
	}
	if authorization.String() == "" {
		http.Error(w, error_msgs.JsonifyError(error_msgs.UNAUTHORIZED), http.StatusUnauthorized)
		return errors.New(error_msgs.UNAUTHORIZED)
	}
	return nil
}
