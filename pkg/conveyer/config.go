package conveyer

//Configuration conveyer
type Config struct {
	// Key and value cookies
	Cookies map[string]string
	// Proxy addr: 127.0.0.1:9050
	Proxy string
	// Timeout for connection
	Timeout int `toml:"timeout"`
	// Delay between passage
	Delay    int `toml:"delay"`
	logLevel string
	// Templates for conveyer steps
	Steps [][]string
}

//Create new config for conveyer:
//  default: &Config{ Steps[0]["download"] }
func NewConfig() *Config {
	steps := make([][]string, 0)
	steps = append(steps, []string{"download"})
	return &Config{
		Steps: steps,
	}
}
