package resources

func Migrations() []string {
	return []string{
		`CREATE TABLE version (version int default 0)`,

		`INSERT INTO version values (0)`,

		`CREATE TABLE config (config TEXT)`,

		`INSERT INTO config values ('{}');`,

		// Search data
		`CREATE TABLE search_items (
title TEXT NOT NULL PRIMARY KEY,
type TEXT NOT NULL CHECK(type in ('Movie', 'Episode', 'Season')),
imdb TEXT,

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
rating INTEGER DEFAULT 0)`,

		`CREATE TABLE downloads (
title TEXT REFERENCES search_items(title) ON DELETE CASCADE,
url TEXT NOT NULL UNIQUE,
destination TEXT NOT NULL,
status TEXT NOT NULL CHECK(status in ('Error', 'Downloading', 'Downloaded')))`,

		// Telegram
		`CREATE TABLE telegram (
id TEXT NOT NULL PRIMARY KEY)`,
	}
}
