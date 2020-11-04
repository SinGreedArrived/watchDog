package main

import (
	"flag"
	"fmt"
	"projects/parser/internal/configure"
	"projects/parser/internal/conveyer"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	CONFIG_PATH string
	logger      *logrus.Logger
	wg          sync.WaitGroup
)

func init() {
	flag.StringVar(&CONFIG_PATH, "config", "configs/config.toml", "path to config file")
	logger = logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{})
}

type pipe interface {
	GetInputChan() chan string
	GetOutputChan() chan []byte
	Start(*sync.WaitGroup) chan string
	Close()
}

func main() {
	flag.Parse()
	config := configure.NewConfig()
	err := config.LoadToml(CONFIG_PATH)
	if err != nil {
		logger.Fatal(err)
	}
	conveyers := make(map[string]pipe)
	for name, configConveyer := range config.GetConveyerConfig() {
		tmp := conveyer.New(name, configConveyer)
		conveyers[name] = tmp
		conveyers[name].Start(&wg)
	}
	list := config.GetTargetList("remanga.org")
	for i := 0; i < 3; i++ {
		conveyers["remanga.org"].GetInputChan() <- list[i]
		fmt.Println(<-conveyers["remanga.org"].GetOutputChan())
	}
	for name, _ := range config.GetConveyerConfig() {
		conveyers[name].Close()
	}

	wg.Wait()
}
