package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"sync"

	"golang.org/x/net/xsrftoken"

	"github.com/gorilla/securecookie"
)

var (
	xsrfkey    = hex.EncodeToString(mustRand(32))
	onlyclient = struct {
		sync.Mutex
		Value string
	}{}
)

func mustRand(n int) []byte {
	if b := securecookie.GenerateRandomKey(n); b != nil {
		return b
	}
	panic("Broken random generator")
}

// onlyOne returns a handler that sets a session cookie on the first connection
// and refuses everybody else. It also protects against xsrf.
// Since the browser ignores port when sending cookies we need to add it to the
// name of the session cookie. this will mess up any XSRF-TOKEN's we'll have set
// talking to other localhost:* servers.
func onlyOne(h http.Handler, port string) http.HandlerFunc {
	cookieName := fmt.Sprintf("SESSION-%s", port)

	return func(w http.ResponseWriter, r *http.Request) {

		onlyclient.Lock()
		if onlyclient.Value == "" {
			onlyclient.Value = hex.EncodeToString(mustRand(32))
			onlyclient.Unlock()

			log.Printf("Session cookie set to %q", onlyclient.Value)
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Path:     "/",
				Value:    onlyclient.Value,
				HttpOnly: true, // not accessible to JS in the browser
			})
			http.SetCookie(w, &http.Cookie{
				Name:     "XSRF-TOKEN",
				Path:     "/",
				Value:    xsrftoken.Generate(xsrfkey, xsrfkey, "use"),
				HttpOnly: false, // yes accessible to JS in the browser, provided same origin

			})

		} else {
			if c, _ := r.Cookie(cookieName); c == nil || c.Value != onlyclient.Value {
				onlyclient.Unlock()
				http.Error(w, fmt.Sprintf("Missing or invalid session cookie."), http.StatusUnauthorized)
				return
			}

			onlyclient.Unlock()
		}

		switch r.Method {
		case "GET", "HEAD":
			// nothing
		default:
			// check cookie is valid and cookie == header
			xsrftok, _ := r.Cookie("XSRF-TOKEN")
			if xsrftok == nil {
				http.Error(w, "Missing xsrf cookie", http.StatusUnauthorized)
				return
			}
			if !xsrftoken.Valid(xsrftok.Value, xsrfkey, xsrfkey, "use") {
				http.Error(w, "Invalid xsrf cookie", http.StatusUnauthorized)
				return
			}
			if xsrfhdr := r.Header.Get("X-XSRF-TOKEN"); xsrfhdr != xsrftok.Value {
				http.Error(w, "Invalid or missing xsrf header", http.StatusUnauthorized)
				return
			}
		}
		h.ServeHTTP(w, r)
	}
}
