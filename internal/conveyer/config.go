package conveyer

type Config struct {
	Cookies  map[string]string
	Proxy    string
	Timeout  int `toml:"timeout"`
	Delay    int `toml:"delay"`
	logLevel string
	Steps    [][]string
}

func NewConfig() *Config {
	steps := make([][]string, 0)
	steps = append(steps, []string{"download"})
	return &Config{
		Steps: steps,
	}
}
