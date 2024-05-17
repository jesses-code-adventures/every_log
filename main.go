package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jesses-code-adventures/every_log/db"
	"github.com/jesses-code-adventures/every_log/endpoints"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Llongfile)
	db := db.NewDb(logger)
	defer db.Close()
	mux := http.NewServeMux()
	handler := endpoints.NewServerHandler(&db, logger)
	mux.Handle("/project/{project_id}/key", endpoints.ApiKeyHandler{Db: &db, Logger: logger})
	mux.Handle("/project/{project_id}/invite", endpoints.ProjectInviteHandler{Db: &db, Logger: logger})
	mux.Handle("/", &handler)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
