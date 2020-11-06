package main

import (
	"encoding/hex"
	"flag"
	"log"
	"os"
	"os/exec"
	"projects/parser/internal/configure"
	"projects/parser/internal/conveyer"
	"projects/parser/internal/model"
	"projects/parser/internal/store"
	"sync"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

var (
	CONFIG_PATH string
	BROWSER     string
	OPEN_NEW    bool
	CLEAR_NEW   bool
	DELAY_OPEN  int
	logger      *logrus.Logger
	wg          sync.WaitGroup
)

func init() {
	flag.StringVar(&CONFIG_PATH, "config", "configs/config.toml", "path to config file")
	flag.StringVar(&BROWSER, "browser", "firefox", "browser for open link")
	flag.BoolVar(&OPEN_NEW, "open", false, "browser for open link")
	flag.BoolVar(&CLEAR_NEW, "clear", false, "clear news")
	flag.IntVar(&DELAY_OPEN, "delay", 1500, "delay for open link")
	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FieldsOrder:     []string{"component", "name"},
		HideKeys:        true,
	})
}

func Collector(s *store.Store, collectChan chan *model.Target) {
	for t := range collectChan {
		elem, _ := s.Target().FindByUrl(t.Url)
		if elem == nil {
			s.Target().Create(t)
			s.News().Create(&model.News{Url: t.Url})
		} else {
			if elem.Hash != t.Hash {
				s.Target().Create(t)
				s.News().Create(&model.News{Url: t.Url})
			} else {
				continue
			}
		}
	}
}

func OpenLink(url string) {
	cmd := exec.Command(BROWSER, url)
	if err := cmd.Start(); err != nil {
		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	logger.Info("Start factory...")
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
	logger.WithField("component", "config").Info("Loaded configure from " + CONFIG_PATH)
	db := store.New(config.Store)
	if err := db.Open(); err != nil {
		logger.WithFields(logrus.Fields{
			"package":  "store",
			"function": "Open",
		}).Fatal(err)
	}
	defer db.Close()
	logger.WithField("component", "store").Info("Opened store from " + config.Store.DBpath)
	if OPEN_NEW {
		urls, _ := db.News().GetAll()
		if len(urls) == 0 {
			logger.WithField("component", "store").Info("I don't have news")
		}
		for _, url := range urls {
			OpenLink(url.Url)
			time.Sleep(time.Millisecond * time.Duration(DELAY_OPEN))
		}
		os.Exit(0)
	}
	if CLEAR_NEW {
		db.News().DeleteAll()
		logger.WithField("component", "store").Info("Clear all news")
		os.Exit(0)
	}
	collectChan := make(chan *model.Target)
	conveyers := make(map[string]*conveyer.Conveyer)
	for name, configConveyer := range config.GetConveyerConfig() {
		conveyers[name], err = conveyer.New(name, configConveyer)
		logger.WithFields(logrus.Fields{
			"component": "conveyer:" + name,
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
	go Collector(db, collectChan)
	for name, conv := range conveyers {
		TargetList := config.GetTargetList(name)
		go func(conv *conveyer.Conveyer, list []string) {
			count := 0
			logger.WithFields(logrus.Fields{
				"component": "conveyer:" + conv.GetName(),
				"status":    "Received",
			}).Infof("%d links", len(TargetList))
			for _, url := range list {
				conv.GetInput() <- []byte(url)
				count++
			}
			conv.Close()
			logger.WithFields(logrus.Fields{
				"component": "conveyer:" + conv.GetName(),
				"status":    "Done",
			}).Infof("%d links", count)
		}(conv, TargetList)
		go func(conv *conveyer.Conveyer, list []string) {
			for _, url := range list {
				hash := <-conv.GetOutput()
				strHash := hex.EncodeToString(hash)
				collectChan <- &model.Target{
					Url:  url,
					Hash: strHash,
				}
			}
		}(conv, TargetList)
	}
	wg.Wait()
	elems, err := db.News().GetAll()
	logger.WithField("component", "store").Infof("News:\n")
	for _, v := range elems {
		logger.WithField("component", "store").Infof("%s:\n", v.Url)
	}
}
