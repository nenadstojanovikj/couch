package download

type MagnetDownloader interface {
	Download(magnet string)
}

