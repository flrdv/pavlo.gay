package main

import (
	"flag"
	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/router/inbuilt/middleware"
	"html/template"
	"log"
	"math"
	"strings"
)

const (
	defaultAddr     = ":80"
	homeTmpl        = "home"
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
	Tmpl *template.Template
}

func (i Index) Handler(request *http.Request) *http.Response {
	name, _ := request.Query.Get("name")
	if len(name) == 0 {
		name = homeDefaultName
	}

	resp := request.Respond()
	if err := i.Tmpl.Execute(resp, name); err != nil {
		return http.Error(request, err)
	}

	return resp
}

func main() {
	flag.Parse()

	tmpl, err := template.ParseFiles(homeTmplPath)
	if err != nil {
		log.Fatalf("cannot load home template: %s", err)
	}

	index := Index{
		Tmpl: tmpl,
	}

	r := inbuilt.New().
		Use(middleware.LogRequests()).
		Use(middleware.Recover).
		Get("/", index.Handler).
		Static("/static", "static").
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
