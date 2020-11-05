package store

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db         *sql.DB
	config     *Config
	TargetRepo *targetRepo
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

	self.db, _ = sql.Open(self.config.db_type, self.config.path)
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
