package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lvdlvd/go-net-http-tmpl"

	git "github.com/libgit2/git2go"
)

var (
	verbose  = flag.Bool("debug", false, "be extra verbose")
	refpfx   = flag.String("ref", "refs/notes/scrutinize", "Notes ref prefix to store review messages on.")
	webroot  = flag.String("webroot", filepath.Join(findHome(), "s"), "Path to dir with static webpages.")
	tmplroot = flag.String("tmplroot", filepath.Join(findHome(), "t"), "Path to dir with template webpages.")
)

var binHome string

func findHome() string {
	if binHome == "" {
		binHome = filepath.Dir(os.Args[0])
		// if we're running from a GOPATH/bin, home is src/github/lvdlvd/git-scrutinizer
		for _, v := range filepath.SplitList(os.Getenv("GOPATH")) {
			if binHome == filepath.Join(v, "bin") {
				binHome = filepath.Join(v, "src", "github.com", "lvdlvd", "git-scrutinizer")
				break
			}
		}
	}
	return binHome
}

var repository *git.Repository

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: git-scrutinize [options] [repo]")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	var (
		repo string
		err  error
	)

	switch len(flag.Args()) {
	case 0:
		repo, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	case 1:
		repo = flag.Arg(0)
	default:
		flag.Usage()
	}

	repository, err = git.OpenRepositoryExtended(repo, 0, "/")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Git repository", repository.Path())

	r := mux.NewRouter()
	r.KeepContext = true // cleared in loghandler

	th := tmpl.NewHandler(filepath.Join(*tmplroot, "*.html"), nil, tmplFuncs)

	r.Path("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/commits", http.StatusMovedPermanently)
	})

	r.Path("/commits").Handler(substPath("commits.html", th))
	r.Path("/tree").Handler(substPath("tree.html", th))
	r.Path("/threads").Handler(substPath("threads.html", th))
	r.Path("/settings").Handler(substPath("settings.html", th))

	exit := make(chan bool)
	r.Path("/quit").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Bye...")
		close(exit)
	})

	// debug/vars and debug/pprof are served by the default mux
	r.PathPrefix("/debug/").Handler(http.DefaultServeMux)

	// all paths that haven't been matched will be served as static files out of the webroot.
	r.Methods("GET", "HEAD").Handler(http.FileServer(http.Dir(*webroot)))

	// :0 lets the OS choose a port
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Opening http://%s", ln.Addr())

	go (&http.Server{
		Addr:    fmt.Sprint(ln.Addr()),
		Handler: logHandler(onlyOne(r, strings.Split(ln.Addr().String(), ":")[1]), *verbose),
	}).Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})

	b, err := exec.Command(openCmd, fmt.Sprintf("http://%s", ln.Addr())).CombinedOutput()
	if len(b) > 0 {
		log.Println(string(b))
	}
	if err != nil {
		log.Fatal(err)
	}

	<-exit
	log.Println("Exiting.")

}

// Invoke h after setting request path to path.
// Saves the original path in mux var "path".
// This is useful because the template handler looks at the url
// path to invoke the template but we want to register under a path without the html.
// TODO automate?
func substPath(p string, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mux.Vars(r)["path"] = r.URL.Path
		r.URL.Path = p
		h.ServeHTTP(w, r)
	}
}

// helper copied from golang.org/pkg/http
type tcpKeepAliveListener struct{ *net.TCPListener }

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	c, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	c.SetKeepAlive(true)
	c.SetKeepAlivePeriod(3 * time.Minute)
	return c, nil
}
