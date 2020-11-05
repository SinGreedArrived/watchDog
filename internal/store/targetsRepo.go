package store

import (
	"projects/parser/internal/model"
)

type targetRepo struct {
	store *Store
}

// Create ...
func (self *targetRepo) Create(t *model.Target) (*model.Target, error) {
	if _, err := self.store.db.Exec(
		"INSERT or REPLACE INTO targets (url,hash) values (?,?)",
		t.Url,
		t.Hash,
	); err != nil {
		return nil, err
	}
	return t, nil
}

// FindByURL ...
func (self *targetRepo) FindByUrl(url string) (*model.Target, error) {
	t := &model.Target{}
	if err := self.store.db.QueryRow(
		"SELECT url, hash FROM targets",
	).Scan(
		&t.Url,
		&t.Hash,
	); err != nil {
		return nil, err
	}
	return t, nil
}

func (self *targetRepo) DeleteByUrl(url string) error {
	if _, err := self.store.db.Exec("DELETE FROM targets WHERE url=?", url); err != nil {
		return err
	}
	return nil
}

func (self *targetRepo) Update(t *model.Target) error {
	if _, err := self.store.db.Exec("UPDATE targets SET hash=? WHERE url=?", t.Hash, t.Url); err != nil {
		return err
	}
	return nil
}
