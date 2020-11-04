package store

type Config struct {
	db_type string `toml:"DBtype"`
	path    string `toml:"DBpath"`
}

func NewConfig() *Config {
	return &Config{
		db_type: "sqlite3",
		path:    "database.db",
	}
}
