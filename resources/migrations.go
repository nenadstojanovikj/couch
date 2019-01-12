package resources

func Migrations() []string {
	return []string{
		`CREATE TABLE version (version int default 0)`,

		`INSERT INTO version values (0)`,

		`CREATE TABLE search_items (
title TEXT NOT NULL PRIMARY KEY,
type TEXT NOT NULL CHECK(type in ('Movie', 'Episode')),
created_at datetime NOT NULL,
updated_at datetime NOT NULL,
status TEXT NOT NULL CHECK(status in ('Pending', 'Debrid', 'Downloading', 'Downloaded', 'Error')))`,

		`CREATE TABLE torrents (
title TEXT REFERENCES search_items(title) ON DELETE CASCADE,
url TEXT NOT NULL UNIQUE,
quality TEXT NOT NULL CHECK(quality in ('4K', 'FHD', 'HD', 'SD'))
)`,
		`CREATE TABLE realdebrid (
title TEXT REFERENCES search_items(title),
url TEXT NOT NULL UNIQUE,
error TEXT
)`,
		`CREATE TABLE torrent_files (
title TEXT REFERENCES search_items(title),
magnet TEXT NOT NULL,

url TEXT NOT NULL UNIQUE
)`,
	}
}
