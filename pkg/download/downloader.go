package download

type Info struct {
	Filepath        string
	TotalBytes      int64
	DownloadedBytes int64
}

func (f *Info) Progress() float64 {
	return float64(f.DownloadedBytes) / float64(f.TotalBytes)
}

type Getter interface {
	Get(url string, destination string) (Informer, error)
}

type Informer interface {
	Info() *Info
}
