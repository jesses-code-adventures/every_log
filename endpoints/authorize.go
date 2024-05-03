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
	"github.com/jesses-code-adventures/every_log/error_msgs"
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
		panic("Error loading .env file")
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
			status := error_msgs.GetErrorHttpStatus(err)
			http.Error(w, error_msgs.JsonifyError(err.Error()), status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Authorized"}`))
	default:
		http.Error(w, error_msgs.JsonifyError(error_msgs.UNACCEPTABLE_HTTP_METHOD), http.StatusMethodNotAllowed)
	}
}

// Ensures the user's authentication credentials are correct and returns a JWT token
// The user should include this token in the Authorization header of future requests
func (a AuthorizationHandler) Authorize(w http.ResponseWriter, r *http.Request) error {
	post, err := newIncomingPostDataAuthorize(r)
	if err != nil {
		status := error_msgs.GetErrorHttpStatus(err)
		http.Error(w, error_msgs.JsonifyError(err.Error()), status)
		return err
	}
	tx, err := a.Db.Db.Begin()
	if err != nil {
		fmt.Println(err) // TODO: use a logger
		newErr := errors.New(error_msgs.DATABASE_ERROR)
		status := error_msgs.GetErrorHttpStatus(newErr)
		http.Error(w, error_msgs.JsonifyError(newErr.Error()), status)
		return newErr
	}
	err = a.Db.Authorize(post.UserId, post.Token, tx)
	if err != nil {
		tx.Rollback()
		status := error_msgs.GetErrorHttpStatus(err)
		http.Error(w, error_msgs.JsonifyError(err.Error()), status)
		return err
	}
	claims, err := a.decodeJWT(post.Token)
	if err != nil {
		tx.Rollback()
		status := error_msgs.GetErrorHttpStatus(err)
		http.Error(w, error_msgs.JsonifyError(err.Error()), status)
		return err
	}
	if claims.ExpiresAt.Time.Before(time.Now()) {
		tx.Rollback()
		fmt.Println(error_msgs.EXPIRED_TOKEN) // TODO: use a logger
		err = errors.New(error_msgs.EXPIRED_TOKEN)
		status := error_msgs.GetErrorHttpStatus(err)
		http.Error(w, error_msgs.JsonifyError(err.Error()), status)
		return err
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
			fmt.Println(fmt.Errorf("unexpected signing method: %v", token.Header["alg"])) // TODO: Use a logger
			return nil, errors.New(error_msgs.AUTHENTICATION_PROCESS_ERROR)
		}
		// Return the secret key for verification
		return []byte(a.signing_key), nil
	})
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return nil, errors.New(error_msgs.AUTHORIZATION_PROCESS_ERROR)
	}
	// Check if the token is valid
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New(error_msgs.INVALID_TOKEN)
}

type incomingAuthorizationData struct {
	UserId string `json:"user_id"`
	Token  string `json:"token"`
}

func getTokenFromCookies(r *http.Request, user_id string) *incomingAuthorizationData {
	cookies := r.Cookies()
	var token string
	for _, cookie := range cookies {
		if cookie.Name == "Authorization" {
			token = cookie.Value
		}
	}
	token = strings.TrimPrefix(token, "Bearer: ")
	if token != "" {
		return &incomingAuthorizationData{UserId: user_id, Token: token}
	}
	return nil
}

// takes the token from cookies if it exists, else looks in the body for "token"
func newIncomingPostDataAuthorize(r *http.Request) (incomingAuthorizationData, error) {
	user_id := r.Header.Get("user_id")
	if user_id == "" {
		fmt.Println(error_msgs.USER_ID_REQUIRED) // TODO: use a logger
		return incomingAuthorizationData{}, errors.New(error_msgs.USER_ID_REQUIRED)
	}
	cookieToken := getTokenFromCookies(r, user_id)
	if cookieToken != nil {
		return *cookieToken, nil
	}
	body := r.Body
	defer body.Close()
	var decodedBody struct {
		Token string `json:"token"`
	}
	err := json.NewDecoder(body).Decode(&decodedBody)
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return incomingAuthorizationData{}, errors.New(error_msgs.JSON_PARSING_ERROR)
	}
	return incomingAuthorizationData{UserId: user_id, Token: strings.TrimPrefix(decodedBody.Token, "Bearer: ")}, nil
}
