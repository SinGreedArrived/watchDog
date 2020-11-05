package main

import (
	"flag"
	"projects/parser/internal/configure"
	"projects/parser/internal/conveyer"
	"projects/parser/internal/model"
	"projects/parser/internal/store"
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
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{})
}

func main() {
	flag.Parse()
	config := configure.NewConfig()
	err := config.LoadToml(CONFIG_PATH)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"package":  "configure",
			"function": "LoadToml",
			"argument": CONFIG_PATH,
		}).Fatal(err)
	}
	db := store.New(config.Store)
	if err := db.Open(); err != nil {
		logger.WithFields(logrus.Fields{
			"package":  "store",
			"function": "Open",
		}).Fatal(err)
	}
	defer db.Close()
	conveyers := make(map[string]*conveyer.Conveyer)
	for name, configConveyer := range config.GetConveyerConfig() {
		conveyers[name], err = conveyer.New(name, configConveyer)
		logger.WithFields(logrus.Fields{
			"conveyer": name,
		}).Info("Create conveyer")
		if err != nil {
			logger.WithFields(logrus.Fields{
				"package":  "conveyer",
				"function": "LoadToml",
				"args[1]":  name,
				"args[2]":  configConveyer,
			}).Warn(err)
		}
		conveyers[name].Start(&wg)
	}
	list := config.GetTargetList("remanga.org")
	//fmt.Fprintf(conveyers["remanga.org"], "%s", list[1])
	for name, _ := range config.GetConveyerConfig() {
		conveyers[name].Close()
	}
	_, err = db.Target().Create(&model.Target{
		Url:  string(list[1]),
		Hash: string("awesome"),
	})
	if err != nil {
		logger.WithFields(logrus.Fields{
			"package":  "store",
			"function": "Target().Create()",
			"args":     "Url:'" + string(list[1]) + "' | Hash:'" + string("awesome") + "'",
		}).Warn(err)
	}
	wg.Wait()
}
