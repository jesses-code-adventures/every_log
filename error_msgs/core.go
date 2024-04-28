package error_msgs

import (
	"fmt"
	"net/http"
	"strings"
)

const JSON_PARSING_ERROR = "Json parsing error"
const AUTHENTICATION_PROCESS_ERROR = "Authentication process error"
const AUTHORIZATION_PROCESS_ERROR = "Authorization process error"
const DATABASE_ERROR = "Database error"
const USER_ID_REQUIRED ="User id required"
const API_KEY_REQUIRED = "Api key required"
const USER_TOKEN_REQUIRED = "User token required"
const AUTHORIZATION_TOKEN_REQUIRED = "Authorization token required"
const EMAIL_EXISTS = "Email exists"
const PROJECT_EXISTS = "Project exists"
const USER_EXISTS = "User exists"
const UNACCEPTABLE_HTTP_METHOD = "Unacceptable http method"
const UNAUTHORIZED = "Unauthorized"
const EXPIRED_TOKEN = "Expired token"
const INVALID_TOKEN = "Invalid token"

func GetRequiredMessage(field string) string {
	return fmt.Sprintf("%s is required", field)
}

func GetExistsMessage(field string) string {
	return fmt.Sprintf("%s already exists", field)
}

func GetErrorHttpStatus(e error) int {
	switch e.Error() {
	case USER_ID_REQUIRED, API_KEY_REQUIRED, USER_TOKEN_REQUIRED, AUTHORIZATION_TOKEN_REQUIRED, EXPIRED_TOKEN, INVALID_TOKEN, UNAUTHORIZED:
		return http.StatusUnauthorized
	case USER_EXISTS, EMAIL_EXISTS, PROJECT_EXISTS:
		return http.StatusConflict
	default:
		if strings.HasSuffix(e.Error(), "is required") {
			return http.StatusUnprocessableEntity
		}
		if strings.HasSuffix(e.Error(), "already exists") {
			return http.StatusConflict
		}
	}
	return http.StatusInternalServerError
}

func JsonifyError(err string) string {
	return `{"error": "` + err + `"}`
}
