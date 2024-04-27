package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jesses-code-adventures/every_log/db"
	"github.com/joho/godotenv"
)

const EXPIRATION_MINUTES = 60


type postRequestData struct {
	UserId   string `json:"user_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func createJWT(userID string, email string, password string, secretKey string) (string, error) {
	// Define custom claims
	claims := jwt.MapClaims{
		"user_id":  userID,
		"email":    email,
		"password": password,
		"iat":      time.Now().Unix(), // Issued at
		"exp":      time.Now().Add(time.Minute * EXPIRATION_MINUTES).Unix(),
	}

	// Create a new token with the claims and signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with a secret key
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func newPostRequestData(r *http.Request) (postRequestData, error) {
	user_id := r.Header.Get("user_id")
	if user_id == "" {
		return postRequestData{}, errors.New("user_id is required")
	}
	body := r.Body
	defer body.Close()
	// Extract fields from JSON body
	var decodedBody struct {
		Email string `json:"email"`;
		Password string `json:"password"`;
	}
	err := json.NewDecoder(body).Decode(&decodedBody)
	if err != nil {
		return postRequestData{}, err
	}
	return postRequestData{UserId: user_id, Email: decodedBody.Email, Password: decodedBody.Password}, nil
}

type AuthenticationHandler struct {
	Db *db.Db
	signing_key string
}

func NewAuthenticationHandler(db *db.Db) AuthenticationHandler {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	jwt_signing_key := os.Getenv("JWT_SIGNING_KEY")
	if jwt_signing_key == "" {
		panic("JWT_SIGNING_KEY environment variable not set")
	}
	return AuthenticationHandler{Db: db, signing_key: jwt_signing_key}
}

func (a AuthenticationHandler) checkAuthenticated(r postRequestData) error {
	isAuthenticated, err := a.Db.Authenticate(r.UserId, r.Password)
	if err != nil {
		return errors.New("Db error")
	}
	if !isAuthenticated {
		return errors.New("Unauthorized")
	}
	return nil
}


// Ensures the user's authentication credentials are correct and returns a JWT token
// The JWT token will be stored in the db and returned in the response
// The user should include this token in the Authorization header of future requests
func (a AuthenticationHandler) Authenticate(r postRequestData) ([]byte, error) {
	tx, err := a.Db.Db.Begin()
	if err != nil {
		return nil, err
	}
	err = a.checkAuthenticated(r)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	token, err := createJWT(r.UserId, r.Email, r.Password, a.signing_key)
	updated := a.Db.UpdateUserToken(r.UserId, token, tx)
	if !updated {
		tx.Rollback()
		return []byte{}, errors.New("Db error")
	}
	response := struct {
		Token string `json:"token"`
	}{
		Token: fmt.Sprintf("Bearer: %s", token),
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("failed to convert response to json")
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

func (a AuthenticationHandler) ServeJson(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		post, err := newPostRequestData(r)
		fmt.Println(post)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := a.Authenticate(post)
		if err != nil {
			if err.Error() == "Db error" || err.Error() == "failed to convert to JSON" {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else if err.Error() == "Unauthorized" {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a AuthenticationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
