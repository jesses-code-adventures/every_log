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
	"github.com/jesses-code-adventures/every_log/error_msgs"
	"github.com/joho/godotenv"
)

type AuthenticationHandler struct {
	Db          *db.Db
	signing_key string
}

func NewAuthenticationHandler(db *db.Db) AuthenticationHandler {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	jwt_signing_key := os.Getenv("JWT_SIGNING_KEY")
	if jwt_signing_key == "" {
		panic("JWT_SIGNING_KEY environment variable not set")
	}
	return AuthenticationHandler{Db: db, signing_key: jwt_signing_key}
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

func (a AuthenticationHandler) ServeJson(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		post, err := newIncomingAuthenticationData(r)
		if err != nil {
			status := error_msgs.GetErrorHttpStatus(err)
			http.Error(w, error_msgs.JsonifyError(err.Error()), status)
			return
		}
		resp, cookie, err := a.authenticate(post)
		if err != nil {
			status := error_msgs.GetErrorHttpStatus(err)
			http.Error(w, error_msgs.JsonifyError(err.Error()), status)
			return
		}
		// TODO: test the cookies are being set properly
		http.SetCookie(w, cookie)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	default:
		http.Error(w, error_msgs.JsonifyError(error_msgs.UNACCEPTABLE_HTTP_METHOD), http.StatusMethodNotAllowed)
	}
}

// Ensures the user's authentication credentials are correct and returns a JWT token
// The JWT token will be stored in the db and returned in the response
// The user should include this token in the Authorization header of future requests
func (a AuthenticationHandler) authenticate(r incomingAuthenticationData) ([]byte, *http.Cookie, error) {
	tx, err := a.Db.Db.Begin()
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return nil, nil, errors.New(error_msgs.DATABASE_ERROR)
	}
	err = a.Db.Authenticate(r.UserId, r.Password, tx)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	token, expires, err := createJWT(r.UserId, r.Email, r.Password, a.signing_key)
	err = a.Db.UpdateUserToken(r.UserId, token, tx)
	if err != nil {
		tx.Rollback()
		return []byte{}, nil, err
	}
	response := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		tx.Rollback()
		fmt.Println(err) // TODO: Use a logger
		return nil, nil, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return nil, nil, errors.New(error_msgs.DATABASE_ERROR)
	}
	cookie := getAuthCookie(token, expires)
	return jsonBytes, cookie, nil
}

func getAuthCookie(token string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     "every_log_auth_token",
		Value:    token, // Your session ID or token
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   true,
	}
}

const EXPIRATION_MINUTES = 60

type incomingAuthenticationData struct {
	UserId   string `json:"user_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func createJWT(userID string, email string, password string, secretKey string) (string, time.Time, error) {
	issued := time.Now()
	expires := issued.Add(time.Minute * EXPIRATION_MINUTES)
	claims := jwt.MapClaims{
		"user_id":  userID,
		"email":    email,
		"password": password,
		"iat":      issued.Unix(),
		"exp":      expires.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return "", time.Time{}, errors.New(error_msgs.AUTHENTICATION_PROCESS_ERROR)
	}
	return signedToken, expires, nil
}

func newIncomingAuthenticationData(r *http.Request) (incomingAuthenticationData, error) {
	user_id := r.Header.Get("user_id")
	if user_id == "" {
		fmt.Println("user id not found") // TODO: Use a logger
		return incomingAuthenticationData{}, errors.New(error_msgs.USER_ID_REQUIRED)
	}
	body := r.Body
	defer body.Close()
	// Extract fields from JSON body
	var decodedBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(body).Decode(&decodedBody)
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return incomingAuthenticationData{}, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	return incomingAuthenticationData{UserId: user_id, Email: decodedBody.Email, Password: decodedBody.Password}, nil
}
