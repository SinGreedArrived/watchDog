package store

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db         *sql.DB
	config     *Config
	TargetRepo *targetRepo
	NewsRepo   *newsRepo
}

// new Store ...
func New(config *Config) *Store {
	return &Store{
		config: config,
	}
}

// Open Store ...
func (self *Store) Open() error {
	// ...
	sql_table := `
CREATE TABLE IF NOT EXISTS "targets" (
	"url"	TEXT UNIQUE,
	"hash"	text,
	PRIMARY KEY("url")
);
CREATE TABLE IF NOT EXISTS "news" (
	"url"	TEXT UNIQUE,
	PRIMARY KEY("url")
);`

	self.db, _ = sql.Open(self.config.DBtype, self.config.DBpath)
	if self.db == nil {
		panic("db nil")
	}
	if err := self.db.Ping(); err != nil {
		return err
	}
	self.db.Exec(sql_table)
	return nil
}

// Close Store ...
func (self *Store) Close() {
	self.db.Close()
}

func (self *Store) Target() *targetRepo {
	if self.TargetRepo != nil {
		return self.TargetRepo
	}
	self.TargetRepo = &targetRepo{
		store: self,
	}
	return self.TargetRepo
}

func (self *Store) News() *newsRepo {
	if self.NewsRepo != nil {
		return self.NewsRepo
	}
	self.NewsRepo = &newsRepo{
		store: self,
	}
	return self.NewsRepo
}
