package store

type Store struct {
	config *Config
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
	return nil
}

// Close Store ...
func (self *Store) Close() {
	// ...
}
