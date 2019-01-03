package cmd

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/spf13/cobra"
	"html/template"
	"net/http"
	"os"
)

const templateDir = "web/templates/"

func NewServerCommand(config *config.Config) *cobra.Command {
	return &cobra.Command{Use: "web", Run: serve(8080)}
}

func serve(port int) func(command *cobra.Command, args []string) {
	return func(command *cobra.Command, args []string) {
		http.HandleFunc("/", showIndex)
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func showIndex(w http.ResponseWriter, r *http.Request) {
	t := template.New("index")
	t, err := template.ParseFiles(templateDir + "header.html")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	t.Execute(w, struct {
		Name string
	}{Name: "nenad"})
}
