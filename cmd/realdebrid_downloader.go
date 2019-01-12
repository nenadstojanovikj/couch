package cmd

//
// import (
// 	"fmt"
// 	"github.com/nenadstojanovikj/couch/pkg/media"
// 	"github.com/nenadstojanovikj/couch/pkg/rss"
// 	"github.com/nenadstojanovikj/couch/pkg/storage"
// 	"github.com/nenadstojanovikj/rd"
// 	"github.com/nenadstojanovikj/showrss-go"
// 	log "github.com/sirupsen/logrus"
// 	"github.com/spf13/cobra"
// 	"os"
// 	"os/signal"
// 	"sort"
// 	"strings"
// 	"syscall"
// 	"time"
// )
//
// // Stores magnet or torrent data in the table
// func NewRealDebridDownloader(debrid *rd.RealDebrid, repo *storage.MediaItemRepository) *cobra.Command {
// 	return &cobra.Command{
// 		Use: "download",
// 		Run: func(cmd *cobra.Command, args []string) {
//
// 			go func() {
//
//
// 				for item := range episodeChan {
// 					urlInfo, err := debrid.Torrents.AddMagnetLinkSimple(item.magnet)
// 					if err != nil {
// 						fmt.Printf("Failed to add magnet %s: %s\n", item.magnet, err)
// 						continue
// 					}
// 					torrentInfo, err := debrid.Torrents.GetTorrent(urlInfo.ID)
// 					if err != nil {
// 						fmt.Printf("Failed to get info about torrent %s: %s\n", urlInfo.ID, err)
// 						continue
// 					}
//
// 					if torrentInfo.Status != rd.StatusWaitingFiles {
// 						fmt.Printf("Skipping torrent in status: %s\n", torrentInfo.Status)
// 						continue
// 					}
//
// 					fmt.Printf("Starting torrent: %s - Status: %s ", torrentInfo.ID, torrentInfo.Status)
// 					var candidateFiles []rd.File
// 					for _, f := range torrentInfo.Files {
// 						if !checkSuffix(f.Path) {
// 							continue
// 						}
// 						candidateFiles = append(candidateFiles, f)
// 					}
//
// 					if len(candidateFiles) == 0 {
// 						fmt.Println("No files, skipping")
// 						continue
// 					}
//
// 					sort.SliceStable(candidateFiles, func(i, j int) bool {
// 						return candidateFiles[i].Bytes < candidateFiles[j].Bytes
// 					})
//
// 					file := candidateFiles[len(candidateFiles)-1]
//
// 					err = debrid.Torrents.SelectFilesFromTorrent(torrentInfo.ID, []int{file.ID})
// 					if err != nil {
// 						fmt.Printf("Error while selecting files: %s\n", err)
// 					} else {
// 						fmt.Printf("Started download for %s\n", item.title)
// 						err = repo.StoreTVShow(item.title)
// 						if err != nil {
// 							fmt.Println(err)
// 							return
// 						}
// 						torrentChan <- torrentTitle{title: item.title, torrent: urlInfo.ID}
// 					}
// 				}
// 			}()
//
// 			// Run a goroutine that will store the final download URL once the torrent is ready
// 			go func() {
// 				for item := range torrentChan {
// 					// TODO Loop check if the torrent has finished downloading
// 					torrent, err := debrid.Torrents.GetTorrent(item.torrent)
// 					if err != nil {
// 						fmt.Println("Could not get torrent")
// 						continue
// 					}
// 					for _, link := range torrent.Links {
// 						data, err := debrid.Unrestrict.SimpleUnrestrict(link)
// 						if err != nil {
// 							fmt.Printf("Cannot unrestrict link: %s %s\n", link, err)
// 							return
// 						}
//
// 						// Used for forcing HTTPS for download
// 						// Should be moved to downloading logic
// 						if strings.HasPrefix(data.Download, "http://") {
// 							data.Download = strings.Replace(data.Download, "http://", "https://", 1)
// 						}
//
// 						err = repo.AddLinks(item.title, []string{data.Download})
// 						fmt.Printf("Download ready: %s\n", data.Download)
// 					}
// 					err = debrid.Torrents.Delete(torrent.ID)
// 					if err != nil {
// 						fmt.Println(err)
// 						continue
// 					}
// 				}
// 			}()
//
// 			// Wait until it's killed by user
// 			<-stop
// 		},
// 	}
// }
//
// func diffMediaTitles(a []media.MediaTitle, b []media.MediaTitle) (diff []media.MediaTitle) {
// 	m := make(map[media.MediaTitle]bool)
//
// 	for _, item := range b {
// 		m[item] = true
// 	}
//
// 	for _, item := range a {
// 		if _, ok := m[item]; !ok {
// 			diff = append(diff, item)
// 		}
// 	}
// 	return diff
// }
