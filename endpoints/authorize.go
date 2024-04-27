package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jesses-code-adventures/every_log/db"
	"github.com/joho/godotenv"
)

type AuthorizationMiddleware interface {
	Authorize(w http.ResponseWriter, r *http.Request) error
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type AuthorizationHandler struct {
	Db          *db.Db
	signing_key string
}

func NewAuthorizationHandler(db *db.Db) AuthorizationHandler {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	jwt_signing_key := os.Getenv("JWT_SIGNING_KEY")
	if jwt_signing_key == "" {
		panic("JWT_SIGNING_KEY environment variable not set")
	}
	return AuthorizationHandler{Db: db, signing_key: jwt_signing_key}
}

func (a AuthorizationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	switch accept {
	case "application/json":
		a.ServeJson(w, r)
		return
	default:
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
}

func (a AuthorizationHandler) ServeJson(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		err := a.Authorize(w, r)
		if err != nil {
			if err.Error() == "Db error" || err.Error() == "failed to convert to JSON" {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else if err.Error() == "Unauthorized" {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			} else if err.Error() == "token expired" {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Authorized"}`))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Ensures the user's authentication credentials are correct and returns a JWT token
// The user should include this token in the Authorization header of future requests
func (a AuthorizationHandler) Authorize(w http.ResponseWriter, r *http.Request) error {
	post, err := newIncomingPostDataAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	tx, err := a.Db.Db.Begin()
	if err != nil {
		return err
	}
	err = a.checkAuthorized(post)
	if err != nil {
		tx.Rollback()
		return err
	}
	claims, err := a.decodeJWT(post.Token)
	if err != nil {
		tx.Rollback()
		return err
	}
	if claims.ExpiresAt.Time.Before(time.Now()) {
		// Roll back the transaction
		tx.Rollback()
		return errors.New("token expired")
	}
	return nil
}

func (a AuthorizationHandler) checkAuthorized(r incomingAuthorizationData) error {
	isAuthorized, err := a.Db.Authorize(r.UserId, r.Token)
	if err != nil {
		return errors.New("Db error")
	}
	if !isAuthorized {
		return errors.New("Unauthorized")
	}
	return nil
}

type Claims struct {
	UserID               string `json:"user_id"`
	Email                string `json:"email"`
	Password             string `json:"password"`
	jwt.RegisteredClaims        // Includes iat, exp, etc.
}

// Function to decode and verify a JWT
func (a *AuthorizationHandler) decodeJWT(tokenString string) (*Claims, error) {
	// Parse the token with the expected claims structure
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify that the signing method is correct
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key for verification
		return []byte(a.signing_key), nil
	})

	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

type incomingAuthorizationData struct {
	UserId string `json:"user_id"`
	Token  string `json:"token"`
}

// takes the token from cookies if it exists, else looks in the body for "token"
func newIncomingPostDataAuthorize(r *http.Request) (incomingAuthorizationData, error) {
	user_id := r.Header.Get("user_id")
	if user_id == "" {
		return incomingAuthorizationData{}, errors.New("user_id is required")
	}
	cookies := r.Cookies()
	var token string
	for _, cookie := range cookies {
		if cookie.Name == "token" {
			token = cookie.Value
		}
	}
	token = strings.TrimPrefix(token, "Bearer: ")
	if token != "" {
		return incomingAuthorizationData{UserId: user_id, Token: token}, nil
	}
	body := r.Body
	defer body.Close()
	// Extract fields from JSON body
	var decodedBody struct {
		Token string `json:"token"`
	}
	err := json.NewDecoder(body).Decode(&decodedBody)
	if err != nil {
		return incomingAuthorizationData{}, err
	}
	return incomingAuthorizationData{UserId: user_id, Token: strings.TrimPrefix(decodedBody.Token, "Bearer: ")}, nil
}
