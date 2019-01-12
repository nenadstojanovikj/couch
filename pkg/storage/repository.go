package storage

import (
	"database/sql"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"strings"
	"time"
)

const (

	// Possible statuses for the media item
	StatusPending     Status = "Pending"
	StatusDownloading Status = "Downloading"
	StatusDownloaded  Status = "Downloaded"
	StatusError       Status = "Error"

	Quality4K  Quality = "4K"
	QualityFHD Quality = "FHD"
	QualityHD  Quality = "HD"
	QualitySD  Quality = "SD"
)

type (
	// Status is the current status of the item
	Status string

	// Media is the model used for storing the item's information
	Media struct {
		Item media.Item

		CreatedAt time.Time
		UpdatedAt time.Time
		Status    Status
	}

	// Quality is the quality of the media
	Quality string

	VideoType string

	Magnet struct {
		Title     media.Title
		Location  string
		Quality   Quality
		VideoType VideoType
	}

	MediaRepository struct {
		db *sql.DB
	}
)

func NewMediaRepository(db *sql.DB) *MediaRepository {
	return &MediaRepository{db}
}

func (r *MediaRepository) StoreMovie(title media.Title) error {
	return r.StoreItem(media.Item{Title: title, Type: media.TypeMovie})
}

func (r *MediaRepository) StoreTVShow(title media.Title) error {
	return r.StoreItem(media.Item{Title: title, Type: media.TypeEpisode})
}

func (r *MediaRepository) StoreItem(item media.Item) error {
	now := time.Now().UTC().Format(ISO8601)
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		"INSERT OR IGNORE INTO search_items (title, type, created_at, updated_at, status) VALUES (?, ?, ?, ?, ?)",
		item.Title, item.Type, now, now, StatusPending,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MediaRepository) AddLinks(title media.Title, links []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	for _, link := range links {
		_, err := tx.Exec("INSERT OR IGNORE INTO realdebrid (title, url) VALUES (?, ?)", title, link)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MediaRepository) AddTorrent(t Magnet) error {
	_, err := r.db.Exec("INSERT OR IGNORE INTO torrents (title, url, quality) VALUES (?, ?, ?)", t.Title, t.Location, t.Quality)
	return err
}

func (r *MediaRepository) Delete(title string) error {
	_, err := r.db.Exec("DELETE FROM search_items WHERE title = ?", title)
	return err
}

// TODO Return show as well
func (r *MediaRepository) NonStartedTorrents() (torrents []string, err error) {
	query := `SELECT m.title FROM search_items m
JOIN torrents t on t.title = m.title
LEFT JOIN realdebrid l on l.title = t.title AND l.title is NULL
WHERE m.status = 'Pending'
`

	rows, err := r.db.Query(query)
	if err != nil {
		return
	}

	for rows.Next() {
		var t string
		err = rows.Scan(&t)
		if err != nil {
			return
		}
		torrents = append(torrents, t)
	}
	return torrents, nil
}

// TODO Parametrize criteria
func (r *MediaRepository) Fetch(criteria ...string) (items []Media, err error) {
	query := "SELECT * FROM search_items"
	if len(criteria) > 0 {
		query += " WHERE " + strings.Join(criteria, " AND ")
	}

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}

	items = make([]Media, 0)
	for rows.Next() {
		var m Media
		err = rows.Scan(&m.Item.Title, &m.Item.Type, &m.CreatedAt, &m.UpdatedAt, &m.Status)
		if err != nil {
			return nil, err
		}
		items = append(items, m)
	}

	return items, nil
}
