# couch

A tool to download your TV shows and movies.

## Inspiration

Vodafone decides to throttle the bandwidth during the evenings, so I have to resort to downloading
my media locally during the day.

## Prerequisites

- `go`, version 1.11+
- most probably Linux system, as some libraries require `libc`
- (opt.) for cross-compilation, you will need `arm-linux-gnueabihf-gcc` and `arm-linux-gnueabihf-g++`


## Usage

To compile it, just run `make build`. The result is the `couch` executable, which can immediately be run to get a list 
of commands.

- `couch run` runs a web server and polls the providers for new media
- `couch auth trakt` will start auth process to Trakt.tv
- `couch auth realdebrid` will start auth process to Real-Debrid (optional, only if you use HTTP download instead of torrent)

The configuration is created after the first `couch run`. It's a JSON string stored in the `config` table in a SQLite database.
Some options are also editable through the web interface at `localhost:8080`.

The authentication procedures must be started after you have run `couch run` at least once (database must be created).

After modifying the config, you must restart the application for the new settings to take effect.

## How it works

`couch` works in a pipeline fashion. It uses four pipelines (stages) to download a file.

- Polling - polls different sources to get new search items. Currently supported providers are Trakt.tv watchlist and calendar
- Scraping - scrapes different torrent sites to get magnet links. Currently supported torrent site is rarbg.com
- Extracting - extracts relevant files from the torrent file. In this case, only downloaded file will be the video(s).
- Downloading - downloads the extracted file(s) to a given location

Related files:
- `cmd/run.go`
- `pkg/media`
- `pkg/download`
- `pkg/magnet`

## Resilience

If the program crashes during download, or any other procedure it will continue after it is restarted. The progress
status is tracked in the database, and the state machine knows where to restart from.


## Torrent flow vs HTTP flow

If the selected downloader is `torrent`, then `couch` will download the torrent through TCP or uTP depending on how it 
was built. Unfortunately, the torrent package will always keep the folder structure when downloading a single file
from the torrent. This means when a movie or TV show will be downloaded it will have the torrent name as a parent folder.

The `http` flow will push the magnet to [Real-Debrid](http://real-debrid.com), and once it's downloaded on the remote
server, `couch` will start downloading the file to the directory specified in the config.
