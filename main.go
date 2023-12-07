package main

import (
	"context"
	"flag"
	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/router/inbuilt/middleware"
	"github.com/indigo-web/indigo/settings"
	"html/template"
	"log"
	"math"
)

const (
	defaultAddr      = ":80"
	defaultHTTPSPort = 443
	homeTmpl         = "home"
	homeTmplPath     = "templates/index.html"
	homeDefaultName  = "Паша"
)

var (
	addr = flag.String(
		"addr", defaultAddr, "address to bind the application",
	)
	https = flag.Bool("tls", false, "enable HTTPS")
	cert  = flag.String(
		"cert", "", "specify TLS certificate path",
	)
	key = flag.String(
		"key", "", "specify TLS private key path",
	)
	httpsPort = flag.Int(
		"httpsport", 0, "HTTPS port to bind the application",
	)
)

func Index(request *http.Request) *http.Response {
	tmpl, ok := request.Ctx.Value(homeTmpl).(*template.Template)
	if !ok {
		return http.Error(request, status.ErrInternalServerError)
	}

	name, _ := request.Query.Get("name")
	if len(name) == 0 {
		name = homeDefaultName
	}

	resp := request.Respond()
	if err := tmpl.Execute(resp, name); err != nil {
		return http.Error(request, err)
	}

	return resp
}

func main() {
	flag.Parse()

	if *httpsPort < 0 || *httpsPort > math.MaxUint16 {
		log.Fatalf("bad https port: %d", *httpsPort)
	}

	tmpl, err := template.ParseFiles(homeTmplPath)
	if err != nil {
		log.Fatalf("cannot load home template: %s", err)
	}

	r := inbuilt.New().
		Use(middleware.LogRequests()).
		Use(middleware.Recover).
		Get("/", Index, middleware.CustomContext(
			context.WithValue(context.Background(), homeTmpl, tmpl),
		)).
		Static("/static", "static").
		Alias("/age", "/static/age.html")

	s := settings.Default()
	s.TLS.Enable = *https
	s.TLS.Cert = *cert
	s.TLS.Key = *key
	s.TLS.Port = uint16(*httpsPort)

	app := indigo.NewApp(*addr)
	log.Printf("Running on %s\n", *addr)
	if err = app.Serve(r, s); err != nil {
		log.Fatal(err)
	}
}
