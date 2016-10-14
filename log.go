package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"time"

	"github.com/gorilla/context"
)

// InterceptWriter serves to catch the status code served for logging
type interceptWriter struct {
	http.ResponseWriter
	status int
}

func (i *interceptWriter) WriteHeader(status int) {
	i.status = status
	i.ResponseWriter.WriteHeader(status)
}

// logHandler wraps a handler, writing a log line after a request and dumping
// request, response and stacktrace on panic.
func logHandler(h http.Handler, verbose bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer context.Clear(r)
		w = &interceptWriter{w, 200}

		defer func(start time.Time) {
			if verbose || w.(*interceptWriter).status >= 400 {
				log.Println(time.Now().Sub(start), r.RemoteAddr, r.Method, r.Host, r.URL, w.(*interceptWriter).status)
			}
			if x := recover(); x != nil {
				log.Printf("%s %s: Crashed: %v", r.Method, r.URL, x)
				log.Println("Request headers:")
				for k, v := range r.Header {
					log.Println("\t", k, ": ", v)
				}
				log.Println("Session variables")
				for k, v := range context.GetAll(r) {
					log.Println("\t", k, ": ", v)
				}
				log.Println("Response headers")
				for k, v := range w.Header() {
					log.Println("\t", k, ": ", v)
				}

				debug.PrintStack()

				log.Print("-----\n\n")
			}
		}(time.Now())

		h.ServeHTTP(w, r)
	}
}

type dumpWriter struct {
	w      http.ResponseWriter
	status int
	body   io.Writer
}

func (d *dumpWriter) Header() http.Header         { return d.w.Header() }
func (d *dumpWriter) Write(b []byte) (int, error) { return d.body.Write(b) }
func (d *dumpWriter) WriteHeader(status int) {
	d.status = status
	d.w.WriteHeader(status)
}

// helper to wrap around any handler
func dump(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Println("dump request:", err)
		}
		var body bytes.Buffer
		dw := dumpWriter{w, 0, io.MultiWriter(&body, w)}
		start := time.Now()
		h.ServeHTTP(&dw, r)
		log.Println(time.Now().Sub(start), r.Method, r.URL, dw.status)
		log.Println("Request body:")
		os.Stderr.Write(d)
		log.Println()
		log.Println("Response headers")
		for k, v := range w.Header() {
			log.Println("\t", k, ": ", v)
		}
		log.Println("Response body:")
		os.Stderr.Write(body.Bytes())
		log.Println()
	}
}
