package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/nenad/couch/pkg/config"
	"github.com/sirupsen/logrus"
)

const templateDir = "web/templates/"

func NewWebServer(config config.Config, store config.Saver) *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/updateSettings", updateConfig(config, store))
	mux.HandleFunc("/", showIndex(config))

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: mux,
	}
}

func updateConfig(conf config.Config, store config.Saver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(conf); err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(fmt.Sprintf("error occurred: %s", err)))
			return
		}

		if err := store.Save(conf); err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(fmt.Sprintf("error occurred: %s", err)))
			return
		}

		w.WriteHeader(200)
	}
}

func showIndex(conf config.Config) http.HandlerFunc {
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
			Port:            conf.Port,
			TVShowDirectory: conf.TVShowsPath,
			MovieDirectory:  conf.MoviesPath,
			DownloaderType:  conf.Downloader,
		})

		if err != nil {
			logrus.Error(err)
		}

		logrus.Debugf("rendered page")
	}
}
