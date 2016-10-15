package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/libgit2/git2go"
)

func postNote(w http.ResponseWriter, r *http.Request) {
	commit := mux.Vars(r)["commit"]

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	if *verbose {
		log.Println(r.Form)
	}
	id, err := git.NewOid(commit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := gitNoteAppend(id, r.Form.Get("text")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
