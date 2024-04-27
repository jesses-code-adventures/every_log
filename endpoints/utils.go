package endpoints

import (
	"errors"
	"net/http"
)

type EndpointHandler interface {
	ServeJson(w http.ResponseWriter, r *http.Request)
}

func BasicValidateRequest(w http.ResponseWriter, r *http.Request) error {
	accept := r.Header.Get("Accept")
	if accept != "application/json" {
		http.Error(w, "Invalid Accept Header", http.StatusBadRequest)
		return errors.New("Accept")
	}
	apiKey := r.Header.Get("x-api-key")
	authorization := r.Header.Get("Authorization")
	if !(r.Method == http.MethodPost && r.RequestURI == "/user") && apiKey  == "" && authorization == ""{
		http.Error(w, "Missing authorization header and api key", http.StatusUnauthorized)
		return errors.New("Auth")
	}
	return nil
}
