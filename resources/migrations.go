package resources

func Migrations() []string {
	return []string{
		`CREATE TABLE version (version int default 0)`,

		`INSERT INTO version values (0)`,

		// Search data
		`CREATE TABLE search_items (
title TEXT NOT NULL PRIMARY KEY,
orig_title TEXT NOT NULL,
type TEXT NOT NULL CHECK(type in ('Movie', 'Episode')),

-- Type specific data
year INTEGER DEFAULT 0,
season SMALLINT DEFAULT 0,
episode SMALLINT DEFAULT 0,

created_at datetime NOT NULL,
updated_at datetime NOT NULL,
status TEXT NOT NULL CHECK(status in ('Pending', 'Scraped', 'Extracting', 'Downloading', 'Downloaded', 'Error')))`,

		// Scrapers
		`CREATE TABLE torrents (
title TEXT REFERENCES search_items(title) ON DELETE CASCADE,
url TEXT NOT NULL UNIQUE,
quality TEXT NOT NULL CHECK(quality in ('4K', 'FHD', 'HD', 'SD')),
encoding TEXT CHECK (encoding in ('x264', 'x265', 'VC-1', 'XviD')),
size INTEGER NOT NULL,
rating INTEGER DEFAULT 1
)`,

		// Handled by downloaders
		`CREATE TABLE realdebrid (
title TEXT REFERENCES search_items(title) ON DELETE CASCADE,
url TEXT NOT NULL UNIQUE,
destination TEXT NOT NULL,
error TEXT
)`,
		`CREATE TABLE torrent_files (
title TEXT REFERENCES search_items(title),
magnet TEXT NOT NULL,
url TEXT NOT NULL UNIQUE,
destination TEXT NOT NULL,
error TEXT
)`,
	}
}
