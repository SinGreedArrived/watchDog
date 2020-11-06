package store

type Config struct {
	DBtype string `toml:"DBtype"`
	DBpath string `toml:"DBpath"`
}

func NewConfig() *Config {
	return &Config{
		DBtype: "sqlite3",
		DBpath: "database.db",
	}
}
