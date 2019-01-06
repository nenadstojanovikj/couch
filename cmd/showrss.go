package cmd

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/rss"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/nenadstojanovikj/rd"
	"github.com/nenadstojanovikj/showrss-go"
	"github.com/spf13/cobra"
	"os"
	"sort"
	"strings"
)

func NewFetchCommand(personalFeed string, repo *storage.MediaItemRepository, feed *showrss.Client, debrid *rd.RealDebrid) *cobra.Command {
	return &cobra.Command{
		Use: "fetch",
		Run: func(cmd *cobra.Command, args []string) {
			// Periodic loop to fetch and compare new episodes
			// go func() {
			//     time.Sleep(time.Second * 10)
			// }()

			episodes, err := feed.GetPersonalEpisodes(personalFeed)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			episodes = rss.FilterMaxDefinitionEpisodes(episodes)

			allEpisodes := make([]storage.MediaTitle, len(episodes))
			for i, e := range episodes {
				allEpisodes[i] = storage.NewTVShowTitle(e.ShowName, e.Season, e.Episode)
			}

			episodeMagnets := make(map[storage.MediaTitle]string, len(episodes))
			for _, e := range episodes {
				episodeMagnets[storage.NewTVShowTitle(e.ShowName, e.Season, e.Episode)] = e.Magnet
			}

			storedMediaItems, err := repo.Fetch(storage.Resources(allEpisodes...))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			storedMediaTitles := make([]storage.MediaTitle, len(storedMediaItems))
			for i, item := range storedMediaItems {
				storedMediaTitles[i] = item.Title
			}

			// Channel that returns the media title and the magnet URL
			newEpisodes := diffMediaTitles(allEpisodes, storedMediaTitles)

			// Channel that returns the torrent ID and the media title
			torrentMap := make(map[string]storage.MediaTitle, len(newEpisodes))

			// Run a goroutine that will download the torrent in RD for each new episode pushed to the channel
			for _, e := range newEpisodes {
				fmt.Println(episodeMagnets[e])
				urlInfo, err := debrid.Torrents.AddMagnetLinkSimple(episodeMagnets[e])
				if err != nil {
					fmt.Printf("Failed to add magnet %s: %s\n", episodeMagnets[e], err)
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
					fmt.Printf("Started download for %s\n", e)
					torrentMap[urlInfo.ID] = e
					err = repo.StoreTVShow(e)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
			}

			for id := range torrentMap {
				torrent, err := debrid.Torrents.GetTorrent(id)
				if err != nil {
					fmt.Println("Could not get torrent")
					continue
				}
				for _, link := range torrent.Links {
					data, err := debrid.Unrestrict.SimpleUnrestrict(link)
					if err != nil {
						fmt.Printf("Cannot unrestrict link: %s %s\n", link, err)
						return
					}

					// Used for forcing HTTPS for download
					// Should be moved to downloading logic
					if strings.HasPrefix(data.Download, "http://") {
						data.Download = strings.Replace(data.Download, "http://", "https://", 1)
					}

					err = repo.AddLinks(torrentMap[id], []string{data.Download})
					fmt.Printf("Download ready: %s\n", data.Download)
				}
				err = debrid.Torrents.Delete(torrent.ID)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
		},
	}
}

func diffMediaTitles(a []storage.MediaTitle, b []storage.MediaTitle) (diff []storage.MediaTitle) {
	m := make(map[storage.MediaTitle]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return diff
}

var allowedExtensions = map[string]bool{".mp4": true, ".mkv": true, ".avi": true, ".m4v": true, ".wmv": true}

func checkSuffix(path string) bool {
	if len(path) < 5 {
		return false
	}
	_, ok := allowedExtensions[path[len(path)-4:]]
	return ok
}
