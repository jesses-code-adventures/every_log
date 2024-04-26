package main

import (
	// "crypto/ecdsa"
	"errors"
	// "github.com/golang-jwt/jwt/v5"
	"github.com/jesses-code-adventures/every_log/db"
	"net/http"
)

type ServerHandler struct {
	db db.Db
}

// func (s *ServerHandler) GenerateToken() jwt.Token {
// 	t := jwt.NewWithClaims(jwt.SigningMethodES256,
// 		jwt.MapClaims{})
// 	s, err := t.SignedString(key)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return *t
// }

func (s *ServerHandler) BasicValidateRequest(r *http.Request) error {
	accept := r.Header.Get("Accept")
	if accept != "application/json" {
		return errors.New("Accept")
	}
	apiKey := r.Header.Get("x-api-key")
	authorization := r.Header.Get("Authorization")
	if apiKey == "" && authorization == "" {
		return errors.New("Auth")
	}
	return nil
}

func (s *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	basicErr := s.BasicValidateRequest(r)
	if basicErr != nil {
		switch basicErr.Error() {
		case "Auth":
			http.Error(w, "Missing authorization header and api key", http.StatusUnauthorized)
		case "Accept":
			http.Error(w, "Invalid Accept Header", http.StatusBadRequest)
		}
		return
	}
	w.Write([]byte("Hello World"))
}

func main() {
	db := db.NewDb()
	defer db.Close()
	mux := http.NewServeMux()
	handler := ServerHandler{db}
	mux.Handle("/", &handler)
	http.ListenAndServe(":8080", mux)
}
