package main

import (
	"net/http"

	"github.com/jesses-code-adventures/every_log/db"
	"github.com/jesses-code-adventures/every_log/endpoints"
)

type ServerHandler struct {
	db    *db.Db
	user  endpoints.UserHandler
	table endpoints.TableHandler
	check endpoints.CheckHandler
}

func NewServerHandler(db *db.Db) ServerHandler {
	return ServerHandler{db: db, user: endpoints.UserHandler{Db: db}, table: endpoints.TableHandler{Db: db}, check: endpoints.CheckHandler{Db: db}}
}

func (s *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	basicErr := endpoints.BasicValidateRequest(w, r)
	if basicErr != nil {
		return
	}
	uri := r.URL.Path
	accept := r.Header.Get("Accept")
	switch uri {
	case "/":
		w.Write([]byte("Hello World"))
	case "/user":
		switch accept {
		case "application/json":
			s.user.ServeJson(w, r)
			return
		default:
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
	case "/table":
		switch accept {
		case "application/json":
			s.table.ServeJson(w, r)
			return
		default:
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
	case "/check":
		switch accept {
		case "application/json":
			s.check.ServeJson(w, r)
			return
		default:
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
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
