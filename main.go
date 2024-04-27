package main

import (
	"net/http"

	"github.com/jesses-code-adventures/every_log/db"
	"github.com/jesses-code-adventures/every_log/endpoints"
)

type ServerHandler struct {
	db           *db.Db
	user         endpoints.UserHandler
	table        endpoints.TableHandler
	check        endpoints.CheckHandler
	authenticate endpoints.AuthenticationHandler
	authorize    endpoints.AuthorizationHandler
}

func NewServerHandler(db *db.Db) ServerHandler {
	return ServerHandler{
		db: db,
		user: endpoints.UserHandler{Db: db},
		table: endpoints.TableHandler{Db: db},
		check: endpoints.CheckHandler{Db: db},
		authenticate: endpoints.AuthenticationHandler{Db: db},
		authorize: endpoints.AuthorizationHandler{Db: db},
	}
}

func (s *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	basicErr := endpoints.BasicValidateRequest(w, r)
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

func main() {
	db := db.NewDb()
	defer db.Close()
	mux := http.NewServeMux()
	handler := NewServerHandler(&db)
	mux.Handle("/", &handler)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
