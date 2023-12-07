package main

import (
	"context"
	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/router/inbuilt/middleware"
	"html/template"
	"log"
)

const (
	addr            = ":8080"
	homeTmpl        = "home"
	homeTmplPath    = "templates/index.html"
	homeDefaultName = "Паша"
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

	app := indigo.NewApp(addr)
	if err = app.Serve(r); err != nil {
		log.Fatal(err)
	}
}
