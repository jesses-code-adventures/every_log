package endpoints

import (
	"net/http"

	"github.com/jesses-code-adventures/every_log/db"
)

type ServerHandler struct {
	db           *db.Db
	user         UserHandler
	table        TableHandler
	check        CheckHandler
	authenticate AuthenticationHandler
	authorize    AuthorizationHandler
}

func NewServerHandler(db *db.Db) ServerHandler {
	return ServerHandler{
		db:           db,
		user:         UserHandler{Db: db},
		table:        TableHandler{Db: db},
		check:        CheckHandler{Db: db},
		authenticate: AuthenticationHandler{Db: db},
		authorize:    AuthorizationHandler{Db: db},
	}
}

func (s *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	basicErr := BasicValidateRequest(w, r)
	if basicErr != nil {
		return
	}
	path := r.URL.Path
	switch path {
	case "/":
		w.Write([]byte(`Welcome to every_log!
Start by creating a user by sending a POST request to /user. The body of your request should be a JSON object of {"email": $1, "password": $2}. You will receive a user_id
You can create a user by sending a POST request to /user, get a list of tables by sending a GET request to /table, and check if a table exists by sending a GET request to /check.`))
	case "/user":
		s.user.ServeHTTP(w, r)
	case "/table":
		s.table.ServeHTTP(w, r)
	case "/check":
		s.check.ServeHTTP(w, r)
	case "/authenticate":
		s.authenticate.ServeHTTP(w, r)
	case "/authorize":
		s.authorize.ServeHTTP(w, r)
	}
}
