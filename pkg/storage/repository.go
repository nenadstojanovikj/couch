package storage

import (
	"database/sql"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"time"
)

const (
	// Possible statuses for the media item
	StatusPending     Status = "Pending"
	StatusScraped     Status = "Scraped"
	StatusExtracting  Status = "Extracting"
	StatusDownloading Status = "Downloading"
	StatusDownloaded  Status = "Downloaded"
	StatusError       Status = "Error"

	Quality4K  Quality = "4K"
	QualityFHD Quality = "FHD"
	QualityHD  Quality = "HD"
	QualitySD  Quality = "SD"

	EncodingXVID Encoding = "XviD"
	Encodingx264 Encoding = "x264"
	Encodingx265 Encoding = "x265"
	EncodingVC1  Encoding = "VC-1"
)

type (
	// Status is the current status of the item
	Status string

	// Media is the model used for storing the item's information
	Media struct {
		Item media.SearchItem

		CreatedAt time.Time
		UpdatedAt time.Time
		Status    Status
	}

	// Quality is the quality of the media
	Quality string

	Encoding string

	Magnet struct {
		Location string
		Quality  Quality
		Encoding Encoding
		Item     media.SearchItem
		Size     uint64 // Size in bytes
		Rating   int
		Seeders  int
	}

	MediaRepository struct {
		db *sql.DB
	}

	Download struct {
		Location    string
		Destination string
		Item        media.SearchItem
	}
)

func NewMediaRepository(db *sql.DB) *MediaRepository {
	return &MediaRepository{db}
}

func (r *MediaRepository) StoreItem(item media.SearchItem) error {
	now := time.Now().UTC().Format(ISO8601)
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		"INSERT OR IGNORE INTO search_items (title, type, created_at, updated_at, status) VALUES (?, ?, ?, ?, ?)",
		item.UniqueTitle, item.Type, now, now, StatusPending,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MediaRepository) AddDownload(download Download) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"INSERT OR IGNORE INTO realdebrid (title, url, destination) VALUES (?, ?, ?)",
		download.Item.UniqueTitle,
		download.Location,
		download.Destination,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MediaRepository) AddTorrent(t Magnet) error {
	_, err := r.db.Exec(`INSERT OR IGNORE INTO torrents (title, url, quality, encoding, rating, size) VALUES (?, ?, ?, ?, ?, ?)`,
		t.Item.UniqueTitle, t.Location, t.Quality, t.Encoding, t.Rating, t.Size)

	return err
}

func (r *MediaRepository) Delete(title string) error {
	_, err := r.db.Exec("DELETE FROM search_items WHERE title = ?", title)
	return err
}

func (r *MediaRepository) Fetch(title string) (m Media, err error) {
	row := r.db.QueryRow("SELECT title, type, status, created_at, updated_at FROM search_items WHERE title = ?", title)
	err = row.Scan(&m.Item.UniqueTitle, &m.Item.Type, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

func (r *MediaRepository) Status(title string, status Status) error {
	_, err := r.db.Exec("UPDATE search_items SET status = ? WHERE title = ?", status, title)
	return err
}

func (r *MediaRepository) InProgressDownloads() (downloads []Download, err error) {
	query := `SELECT m.title, m.type, l.url, l.destination FROM search_items m
JOIN realdebrid l on l.title = m.title
WHERE m.status in ('Extracting', 'Downloading')
`

	rows, err := r.db.Query(query)
	if err != nil {
		return
	}

	for rows.Next() {
		var d Download
		err = rows.Scan(&d.Item.UniqueTitle, &d.Item.Type, &d.Location, &d.Destination)
		if err != nil {
			return
		}
		downloads = append(downloads, d)
	}
	return downloads, nil
}

func (r *MediaRepository) NonExtractedTorrents() (torrents []Magnet, err error) {
	query := `SELECT m.title, m.type, t.url, t.size, t.quality, t.encoding, t.rating FROM search_items m
JOIN torrents t on t.title = m.title
-- JOIN realdebrid l on l.title = t.title AND l.title is NULL
WHERE m.status in ('Extracting', 'Scraped')
AND m.title NOT IN (SELECT l.title FROM realdebrid l)
GROUP BY t.title
ORDER BY t.rating ASC;
`

	rows, err := r.db.Query(query)
	if err != nil {
		return
	}

	for rows.Next() {
		var t Magnet
		err = rows.Scan(&t.Item.UniqueTitle, &t.Item.Type, &t.Location, &t.Size, &t.Quality, &t.Encoding, &t.Rating)
		if err != nil {
			return
		}
		torrents = append(torrents, t)
	}
	return torrents, nil
}
