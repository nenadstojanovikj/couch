package resources

func Migrations() []string {
	return []string{
		`CREATE TABLE version (version int default 0)`,

		`INSERT INTO version values (0)`,

		`CREATE TABLE resources (
title TEXT NOT NULL PRIMARY KEY,
res_type TEXT NOT NULL CHECK(res_type in ('Movie', 'TVShow')),
created_at datetime NOT NULL,
updated_at datetime NOT NULL,
status TEXT NOT NULL CHECK(status in ('Pending', 'Downloading', 'Downloaded', 'Error')))`,

		`CREATE TABLE links (
title TEXT REFERENCES resources(title) ON DELETE CASCADE,
url TEXT NOT NULL UNIQUE,
error TEXT
)`,
	}
}
