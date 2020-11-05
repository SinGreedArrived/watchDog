CREATE TABLE IF NOT EXISTS "targets" (
	"url"	TEXT UNIQUE,
	"hash"	text,
	PRIMARY KEY("url")
);

CREATE TABLE IF NOT EXISTS "news" (
	"url"	TEXT UNIQUE,
	PRIMARY KEY("url")
);`
