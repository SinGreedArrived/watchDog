package configure

import (
	"io/ioutil"
	"os"
	"projects/parser/internal/conveyer"
	"projects/parser/internal/store"

	"github.com/pelletier/go-toml"
)

type config struct {
	Conveyer map[string]*conveyer.Config `toml:"conveyer"`
	Targets  map[string][]string         `toml:"targets"`
	Store    *store.Config               `toml:"store"`
}

func NewConfig() *config {
	return &config{
		Store: store.NewConfig(),
	}
}

func (self *config) LoadToml(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	if err := toml.Unmarshal(data, self); err != nil {
		return err
	}
	return nil
}

func (self *config) GetConv(name string) ([]byte,error) {
	data,err := toml.Marshal(self.Conveyer[name])
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (self *config) GetConveyerConfig() map[string]*conveyer.Config {
	return self.Conveyer
}

func (self *config) GetTargetList(sourceName string) []string {
	return self.Targets[sourceName]
}
