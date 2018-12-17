package main

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/cmd"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"os"
)

func main() {
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

	// token := rd.Token{}
	// if conf.RealDebrid.RefreshToken != "" {
	//     token.AccessToken = conf.RealDebrid.AccessToken
	//     token.RefreshToken = conf.RealDebrid.RefreshToken
	//     token.ExpiresIn = conf.RealDebrid.ExpiresIn
	//     token.ObtainedAt = conf.RealDebrid.ObtainedAt
	//     token.TokenType = conf.RealDebrid.TokenType
	// }

	// rdc := rd.NewRealDebrid(token, http.DefaultClient, rd.AutoRefresh)

	// torr, err := rdc.Torrents.GetTorrent("YDGPTUZOLNBNO")
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	//
	// fmt.Println(torr)
}
