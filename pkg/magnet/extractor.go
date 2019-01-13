package magnet

var allowedExtensions = map[string]bool{".mp4": true, ".mkv": true, ".avi": true, ".m4v": true, ".wmv": true}

type Extractor interface {
	Extract(magnet string) (string, error)
}
