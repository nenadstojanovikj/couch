package web

import (
	"encoding/json"
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/sirupsen/logrus"
	"html/template"
	"net/http"
)

const templateDir = "web/templates/"

func NewWebServer(config *config.Config) *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/updateSettings", updateConfig(config))
	mux.HandleFunc("/", showIndex(config))

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: mux,
	}
}

func updateConfig(config *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(config); err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(fmt.Sprintf("error occurred: %s", err)))
			return
		}

		if err := config.Save(); err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(fmt.Sprintf("error occurred: %s", err)))
			return
		}

		w.WriteHeader(200)
	}
}

func showIndex(config *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := template.New("main")
		t, err := template.ParseGlob(templateDir + "*")
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		err = t.ExecuteTemplate(w, "settings", struct {
			Port            int
			MovieDirectory  string
			TVShowDirectory string
			DownloaderType  string
		}{
			Port:            config.Port,
			TVShowDirectory: config.TVShowsPath,
			MovieDirectory:  config.MoviesPath,
			DownloaderType:  config.Downloader,
		})

		if err != nil {
			logrus.Error(err)
		}

		logrus.Debugf("rendered page")
	}
}
