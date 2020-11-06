package store

import (
	"fmt"
	"testing"
)

func TestStore(t *testing.T, db_type string, db_path string) (*Store, func(...string)) {
	t.Helper()
	config := NewConfig()
	config.DBtype = "sqlite3"
	config.DBpath = "database_test"
	s := New(config)
	if err := s.Open(); err != nil {
		t.Fatal(err)
	}
	return s, func(tables ...string) {
		if len(tables) > 0 {
			if _, err := s.db.Exec(fmt.Sprintf("drop table %s", tables)); err != nil {
				t.Fatal(err)
			}
		}
		s.Close()
	}
}
