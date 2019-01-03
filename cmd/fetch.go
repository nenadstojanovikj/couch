package cmd

import (
	"fmt"
	"github.com/cavaliercoder/grab"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/couch/pkg/rss"
	"github.com/nenadstojanovikj/rd"
	"github.com/nenadstojanovikj/showrss-go"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

func NewFetchCommand(conf *config.Config) *cobra.Command {
	return &cobra.Command{
		Use: "fetch",
		Run: fetch(conf),
	}
}

func fetch(conf *config.Config) func(cmd *cobra.Command, args []string) {
	return func(command *cobra.Command, args []string) {
		feed := showrss.NewClient(http.DefaultClient)
		episodes, err := feed.GetPersonalEpisodes(conf.ShowRss.PersonalFeed)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		episodes = rss.FilterMaxDefinitionEpisodes(episodes)
		if len(episodes) == 0 {
			return
		}

		debrid := rd.NewRealDebrid(createToken(&conf.RealDebrid), http.DefaultClient, rd.AutoRefresh)

		for _, e := range episodes {
			fmt.Println(e.Magnet)
			urlInfo, err := debrid.Torrents.AddMagnetLinkSimple(e.Magnet)
			if err != nil {
				fmt.Printf("Failed to add magnet %s: %s\n", e.Magnet, err)
				continue
			}
			torrentInfo, err := debrid.Torrents.GetTorrent(urlInfo.ID)
			if err != nil {
				fmt.Printf("Failed to get info about torrent %s: %s\n", urlInfo.ID, err)
				continue
			}

			if torrentInfo.Status != rd.StatusWaitingFiles {
				fmt.Printf("Skipping torrent in status: %s\n", torrentInfo.Status)
				continue
			}

			fmt.Printf("Starting torrent: %s - Status: %s ", torrentInfo.ID, torrentInfo.Status)
			var candidateFiles []rd.File
			for _, f := range torrentInfo.Files {
				if !checkSuffix(f.Path) {
					continue
				}
				candidateFiles = append(candidateFiles, f)
			}

			if len(candidateFiles) == 0 {
				fmt.Println("No files, skipping")
				continue
			}

			sort.SliceStable(candidateFiles, func(i, j int) bool {
				return candidateFiles[i].Bytes < candidateFiles[j].Bytes
			})

			file := candidateFiles[len(candidateFiles)-1]

			err = debrid.Torrents.SelectFilesFromTorrent(torrentInfo.ID, []int{file.ID})
			if err != nil {
				fmt.Printf("Error while selecting files: %s\n", err)
			} else {
				fmt.Printf("Started download for %s %dx%d - Quality: %s\n", e.ShowName, e.Season, e.Episode, e.Quality)
				break
			}
		}

		list, err := debrid.Torrents.GetTorrents()
		if err != nil {
			fmt.Printf("Cannot get list: %s", err)
			os.Exit(1)
		}

		grabClient := grab.NewClient()

		var torrent rd.TorrentInfo
		for _, t := range list {
			if t.Status == rd.StatusDownloaded {
				torrent = t
				break
			}
		}

		data, err := debrid.Unrestrict.SimpleUnrestrict(torrent.Links[0])
		if err != nil {
			fmt.Printf("Cannot unrestrict link: %s %s\n", torrent.Links[0], err)
			return
		}

		if strings.HasPrefix(data.Download, "http://") {
			data.Download = strings.Replace(data.Download, "http://", "https://", 1)
		}

		req, err := grab.NewRequest(path.Base(data.Download), data.Download)
		if err != nil {
			fmt.Printf("Cannot unrestrict link: %s %s\n", torrent.Links[0], err)
			return
		}
		transfer := grabClient.Do(req)

		fmt.Println(data.Download)
		go func(transfer *grab.Response) {
			for {
				fmt.Printf("%d/%d - %f (%s)\r", transfer.BytesComplete(), transfer.Size, transfer.Progress(), transfer.ETA())
				time.Sleep(time.Second)
			}
		}(transfer)

		if transfer.Err() != nil {
			fmt.Printf("ERROR: %s", transfer.Err())
		}

		// for _, t := range list {
		// 	if len(t.Links) == 0 {
		// 		continue
		// 	}
		//
		// 	data, err := debrid.Unrestrict.SimpleUnrestrict(t.Links[0])
		// 	if err != nil {
		// 		fmt.Printf("Cannot unrestrict link: %s %s\n", t.Links[0], err)
		// 		continue
		// 	}
		// 	fmt.Printf("Download ready: %s\n", data.Download)
		//
		// }

	}
}

var allowedExtensions = map[string]bool{".mp4": true, ".mkv": true, ".avi": true, ".m4v": true, ".wmv": true}

func checkSuffix(path string) bool {
	if len(path) < 5 {
		return false
	}
	_, ok := allowedExtensions[path[len(path)-4:]]
	return ok
}

func createToken(conf *config.AuthConfig) rd.Token {
	return rd.Token{
		AccessToken:  conf.AccessToken,
		TokenType:    conf.TokenType,
		ExpiresIn:    conf.ExpiresIn,
		ObtainedAt:   conf.ObtainedAt,
		RefreshToken: conf.RefreshToken,
	}
}
