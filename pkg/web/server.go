package web

import (
	"fmt"
	"html/template"
	"net/http"
)

const templateDir = "web/templates/"

func NewWebServer(port int) *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/", showIndex)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
}

func showIndex(w http.ResponseWriter, r *http.Request) {
	t := template.New("index")
	t, err := template.ParseFiles(templateDir + "header.html")
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_ = t.Execute(w, struct {
		Name string
	}{Name: "nenad"})
}
