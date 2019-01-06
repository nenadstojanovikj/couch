package main

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/cmd"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, os.Kill, syscall.SIGTERM)

	conf := config.NewConfiguration()

	if err := conf.LoadFromFile("config.json"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rootCmd := cmd.NewCLI(&conf)
	err := rootCmd.Execute()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Wait until it's killed by user
	<-stop
}
