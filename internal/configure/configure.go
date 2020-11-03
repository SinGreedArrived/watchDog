package configure

import (
	"io/ioutil"
	"log"
	"os"
	"projects/parser/internal/conveyer"

	"github.com/BurntSushi/toml"
)

type config struct {
	Conveyer map[string]*conveyer.Config `toml:"conveyer"`
	Targets  map[string]*targetList      `toml:"targets"`
}

type targetList struct {
	List []string
}

func New() *config {
	return &config{}
}

func (self *config) LoadToml(filename string) error {
	log.Println("Load Toml file")
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	if _, err := toml.Decode(string(data), self); err != nil {
		return err
	}
	return nil
}

func (self *config) GetConveyerConfig(name string) *conveyer.Config {
	return self.Conveyer[name]
}
