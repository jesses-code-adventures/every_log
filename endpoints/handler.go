package endpoints

import (
	"log"
	"net/http"

	"github.com/jesses-code-adventures/every_log/db"
)

type ServerHandler struct {
	db           *db.Db
	user         UserHandler
	table        TableHandler
	check        CheckHandler
	authenticate AuthenticationHandler
	authorize    AuthorizationMiddleware
	project      ProjectHandler
	dbUser       DbUserHandler
	log          LogHandler
	org          OrgHandler
	Logger       *log.Logger
}

func NewServerHandler(db *db.Db, logger *log.Logger) ServerHandler {
	handler := ServerHandler{
		db:           db,
		user:         UserHandler{Db: db, Logger: logger},
		table:        TableHandler{Db: db, Logger: logger},
		check:        CheckHandler{Db: db, Logger: logger},
		authenticate: AuthenticationHandler{Db: db, Logger: logger},
		authorize:    AuthorizationHandler{Db: db, Logger: logger},
		project:      ProjectHandler{Db: db, Logger: logger},
		dbUser:       DbUserHandler{Db: db, Logger: logger},
		log:          LogHandler{Db: db, Logger: logger},
		org:          OrgHandler{Db: db, Logger: logger},
	}
	return handler
}

func (s *ServerHandler) HandleAuthMiddleware(w http.ResponseWriter, r *http.Request, handler http.HandlerFunc) {
	err := s.authorize.Authorize(w, r)
	if err != nil {
		//TODO: should http headers be set here?
		s.Logger.Println("got auth middleware error: ", err)
		return
	}
	handler.ServeHTTP(w, r)
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
	case "/dev_db_user":
		s.dbUser.ServeHTTP(w, r)
	case "/authenticate":
		s.authenticate.ServeHTTP(w, r)
	case "/authorize":
		s.authorize.ServeHTTP(w, r)
	case "/project":
		s.HandleAuthMiddleware(w, r, s.project.ServeHTTP)
	case "/log":
		s.HandleAuthMiddleware(w, r, s.log.ServeHTTP)
	case "/org":
		s.HandleAuthMiddleware(w, r, s.org.ServeHTTP)
	}
}
