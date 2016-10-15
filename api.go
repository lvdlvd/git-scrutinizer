package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func postNote(w http.ResponseWriter, r *http.Request) {
	commit := mux.Vars(r)["commit"]

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	if *verbose {
		log.Println(r.Form)
	}
}
