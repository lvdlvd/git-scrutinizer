package main

import (
	"fmt"
	"log"
	"net/http"
	"net/textproto"

	"github.com/gorilla/mux"

	git "github.com/libgit2/git2go"
)

func postNote(w http.ResponseWriter, r *http.Request) {
	commit := mux.Vars(r)["commit"]
	// mux guarantees this is set, but not that it is valid
	// TODO: move lexical check to mux, validate its a commit here?
	if _, err := git.NewOid(commit); err != nil {
		http.Error(w, fmt.Sprintf("commit is not a valid oid: %v", err), http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if *verbose {
		log.Println(r.Form)
	}
	id, err := git.NewOid(commit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg := Message{
		Header: textproto.MIMEHeader{},
		Body:   r.Form.Get("text"),
	}

	for k, v := range r.Form {
		if k == "text" {
			continue
		}
		msg.Header[textproto.CanonicalMIMEHeaderKey(k)] = v
	}

	if err := gitNoteAppend(id, &msg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
