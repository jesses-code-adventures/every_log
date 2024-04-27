package main

import (
	"net/http"

	"github.com/jesses-code-adventures/every_log/db"
	"github.com/jesses-code-adventures/every_log/endpoints"
)

func main() {
	db := db.NewDb()
	defer db.Close()
	mux := http.NewServeMux()
	handler := endpoints.NewServerHandler(&db)
	mux.Handle("/", &handler)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
