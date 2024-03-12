package main

import (
	"flag"
	"fmt"
	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/router/inbuilt/middleware"
	"html/template"
	"log"
	"math"
	"strings"
	"sync"
)

const (
	defaultAddr     = ":80"
	homeTmplPath    = "templates/index.html"
	homeDefaultName = "Паша"
)

var (
	addr = flag.String(
		"addr", defaultAddr, "address to bind the application",
	)
	https = flag.String(
		"https", "",
		"sets HTTPS up. Default value uses auto-https, otherwise comma-separated "+
			"paths to certificate and key must be provided respectively",
	)
	httpsPort = flag.Int(
		"httpsport", 443, "HTTPS port to bind the application",
	)
)

type Index struct {
	mu   *sync.RWMutex
	tmpl *template.Template
	path string
}

func NewIndex(tmplPath string) (*Index, error) {
	tmpl, err := template.ParseFiles(homeTmplPath)
	if err != nil {
		return nil, fmt.Errorf("cannot load home template: %s", err)
	}

	return &Index{
		mu:   new(sync.RWMutex),
		tmpl: tmpl,
		path: tmplPath,
	}, nil
}

func (i *Index) Render(request *http.Request) *http.Response {
	name, _ := request.Query.Get("name")
	if len(name) == 0 {
		name = homeDefaultName
	}

	resp := request.Respond()
	i.mu.RLock()
	defer i.mu.RUnlock()
	if err := i.tmpl.Execute(resp, name); err != nil {
		return http.Error(request, err)
	}

	return resp
}

func (i *Index) ReloadTemplate(request *http.Request) *http.Response {
	i.mu.Lock()
	defer i.mu.Unlock()

	tmpl, err := template.ParseFiles(homeTmplPath)
	if err != nil {
		return http.Error(request, err)
	}

	i.tmpl = tmpl

	return http.String(request, "reloaded the template successfully")
}

func antiPathTraversal(next inbuilt.Handler, request *http.Request) *http.Response {
	if !isSafe(request.Path) {
		return http.Error(request, status.ErrNotFound)
	}

	return next(request)
}

// isSafe checks for path traversal (basically - double dots)
func isSafe(path string) bool {
	for len(path) > 0 {
		dot := strings.IndexByte(path, '.')
		if dot == -1 {
			return true
		}

		if dot < len(path)-1 && path[dot+1] == '.' {
			return false
		}
	}

	return true
}

func main() {
	flag.Parse()

	index, err := NewIndex(homeTmplPath)
	if err != nil {
		log.Fatalf("index: %s", err)
		return
	}

	r := inbuilt.New().
		Use(middleware.Recover).
		Use(middleware.LogRequests()).
		Get("/", index.Render).
		Get("/reload-template", index.ReloadTemplate).
		Static("/static", "static", antiPathTraversal).
		Alias("/age", "/static/age.html")

	if *httpsPort < 0 || *httpsPort > math.MaxUint16 {
		log.Fatalf("bad https port: %d", *httpsPort)
	}

	app := indigo.New(*addr)

	if len(*https) == 0 {
		app.AutoHTTPS(uint16(*httpsPort))
	} else {
		cert, key := splitPaths(*https)
		app.HTTPS(uint16(*httpsPort), cert, key)
	}

	app.NotifyOnStart(func() {
		log.Printf("Running on %s\n", *addr)
	})

	log.Fatal(app.Serve(r))
}

func splitPaths(paths string) (cert, key string) {
	files := strings.SplitN(paths, ",", 2)
	if len(files) < 2 {
		panic("bad HTTPS cert and key pair")
	}

	return files[0], files[1]
}
