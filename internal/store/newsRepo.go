package store

import "projects/parser/internal/model"

type newsRepo struct {
	store *Store
}

func (self *newsRepo) Create(n *model.News) (*model.News, error) {
	if _, err := self.store.db.Exec(
		"INSERT or REPLACE INTO news (url,open) values (?,?)",
		n.Url,
		n.Open,
	); err != nil {
		return nil, err
	}
	return n, nil
}

func (self *newsRepo) DeleteByUrl(url string) error {
	if _, err := self.store.db.Exec(
		"Delete from news where url=?",
		url,
	); err != nil {
		return err
	}
	return nil
}

func (self *newsRepo) DeleteAll() error {
	if _, err := self.store.db.Exec(
		"delete from news where open=true",
	); err != nil {
		return err
	}
	return nil
}

func (self *newsRepo) GetAll() ([]*model.News, error) {
	result := make([]*model.News, 0)
	rows, err := self.store.db.Query("select url, open from news")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &model.News{}
		if err = rows.Scan(&tmp.Url, &tmp.Open); err != nil {
			return nil, err
		}
		result = append(result, tmp)
	}
	return result, nil
}

func (self *newsRepo) CheckOpened(url string) error {
	if _, err := self.store.db.Exec(
		"update news set open=? where url=?",
		true,
		url,
	); err != nil {
		return err
	}
	return nil
}
