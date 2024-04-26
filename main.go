package main

import (
	"flag"
	"fmt"
	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/router/inbuilt/middleware"
	"html/template"
	"log"
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
		"http", defaultAddr, "specify the server address",
	)
	https = flag.String(
		"https", "",
		"specify the https server address. Leave empty to not use HTTPS at all",
	)
	cert = flag.String(
		"cert", "",
		"specify custom server certificate instead of autocert",
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

func main() {
	flag.Parse()

	index, err := NewIndex(homeTmplPath)
	if err != nil {
		log.Fatalf("parse index template: %s", err)
		return
	}

	r := inbuilt.New().
		Use(middleware.Recover).
		Use(middleware.LogRequests()).
		Get("/", index.Render).
		Get("/reload-template", index.ReloadTemplate).
		Static("/static", "static").
		Alias("/age", "/static/age.html")

	app := indigo.New(*addr)

	if len(*https) > 0 {
		if len(*cert) > 0 {
			certificate, key := splitPaths(*cert)
			app.HTTPS(*https, certificate, key)
		} else {
			app.AutoHTTPS(*https)
		}
	}

	err = app.OnBind(func(addr string) {
		log.Printf("listening on %s\n", addr)
	}).Serve(r)
	log.Fatal(err)
}

func splitPaths(paths string) (cert, key string) {
	files := strings.SplitN(paths, ",", 2)
	if len(files) < 2 {
		panic("bad HTTPS cert and key pair")
	}

	return files[0], files[1]
}
