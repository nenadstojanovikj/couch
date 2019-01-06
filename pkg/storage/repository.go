package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	// Possible types of the media item
	TypeMovie  MediaType = "Movie"
	TypeTVShow MediaType = "TVShow"

	// Possible statuses for the media item
	StatusPending     Status = "Pending"
	StatusDownloading Status = "Downloading"
	StatusDownloaded  Status = "Downloaded"
	StatusError       Status = "Error"

	// Formats for the title of the media item per type
	FormatMovie  string = "%s (%d)"
	FormatTVShow string = "%s (S%02dE%02d)"
)

type (
	// MediaTitle represents the Movie or TVShow title. Should be generated
	// through the NewMovieTitle or NewTVShowTitle package methods
	MediaTitle string

	// MediaType is the type of media
	MediaType string

	// Status is the current status of the item
	Status string

	// MediaItem is the model used for storing the item's information
	MediaItem struct {
		Title MediaTitle
		Type  MediaType

		CreatedAt time.Time
		UpdatedAt time.Time
		Status    Status
	}

	MediaItemRepository struct {
		db *sql.DB
	}
)

func NewMovieTitle(title string, year int) MediaTitle {
	return MediaTitle(fmt.Sprintf(FormatMovie, title, year))
}

func NewTVShowTitle(title string, season, episode int) MediaTitle {
	return MediaTitle(fmt.Sprintf(FormatTVShow, title, season, episode))
}

func NewMediaItemRepository(db *sql.DB) *MediaItemRepository {
	return &MediaItemRepository{db}
}

func (r *MediaItemRepository) StoreMovie(title MediaTitle) error {
	return r.storeMediaItem(title, TypeMovie)
}

func (r *MediaItemRepository) StoreTVShow(title MediaTitle) error {
	return r.storeMediaItem(title, TypeTVShow)
}

func (r *MediaItemRepository) storeMediaItem(title MediaTitle, mediaType MediaType) error {
	now := time.Now().UTC().Format(ISO8601)
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		"REPLACE INTO resources (title, res_type, created_at, updated_at, status) VALUES (?, ?, ?, ?, ?)",
		title, mediaType, now, now, StatusPending,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MediaItemRepository) AddLinks(title MediaTitle, links []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	for _, link := range links {
		_, err := tx.Exec("INSERT OR IGNORE INTO links (title, url) VALUES (?, ?)", title, link)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MediaItemRepository) Delete(title string) error {
	_, err := r.db.Exec("DELETE FROM resources WHERE title = ?", title)
	return err
}

func (r *MediaItemRepository) Fetch(criteria ...string) (resources []MediaItem, err error) {
	query := "SELECT * FROM resources"
	if len(criteria) > 0 {
		query += " WHERE " + strings.Join(criteria, " AND ")
	}

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}

	resources = make([]MediaItem, 0)
	for rows.Next() {
		var res MediaItem
		err = rows.Scan(&res.Title, &res.Type, &res.CreatedAt, &res.UpdatedAt, &res.Status)
		if err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}

	return resources, nil
}
