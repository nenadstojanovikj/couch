package extractor

type Extractor interface {
	Extract(magnet string) (string, error)
}
