package main

import (
	"context"
	"encoding/hex"
	"flag"
	"os"
	"os/exec"
	"os/signal"
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
		HideKeys:        true,
	})
}

func Collector(ctx context.Context, s *store.Store, collectChan chan *model.Target) {
	for {
		select {
		case t := <-collectChan:
			elem, _ := s.Target().FindByUrl(t.Url)
			if elem == nil {
				_, err := s.Target().Create(t)
				if err != nil {
					logger.Panic(err)
				}
				_, err = s.News().Create(&model.News{Url: t.Url, Open: false})
				if err != nil {
					logger.Panic(err)
				}
				logger.WithFields(logrus.Fields{
					"component": "Store",
					"table":     "News",
				}).Infof("Create %s", t.Url)

			} else {
				if elem.Hash != t.Hash {
					_, err := s.Target().Create(t)
					if err != nil {
						logger.Panic(err)
					}
					_, err = s.News().Create(&model.News{Url: t.Url, Open: false})
					if err != nil {
						logger.Panic(err)
					}
					logger.WithFields(logrus.Fields{
						"component": "Store",
						"table":     "News",
					}).Infof("Create %s", t.Url)
				} else {
					continue
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func OpenLink(url string) {
	cmd := exec.Command(BROWSER, url)
	if err := cmd.Start(); err != nil {
		if err != nil {
			logger.Error(err)
		}
	}
}

func handleSignals(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	for {
		sig := <-sigCh
		switch sig {
		case os.Interrupt:
			logger.Error("Signal Interrupt!")
			cancel()
			return
		}
	}
}

func main() {
	logger.Info("Start factory...")
	flag.Parse()
	ctx, cancelFunc := context.WithCancel(context.Background())
	go handleSignals(cancelFunc)
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
			"package":  "Store",
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
			db.News().CheckOpened(url.Url)
			time.Sleep(time.Millisecond * time.Duration(DELAY_OPEN))
		}
		os.Exit(0)
	}
	if CLEAR_NEW {
		db.News().DeleteAll()
		logger.WithField("component", "store").Info("Deleted opened news")
		os.Exit(0)
	}
	collectChan := make(chan *model.Target)
	conveyers := make(map[string]*conveyer.Conveyer)
	for name, configConveyer := range config.GetConveyerConfig() {
		conveyers[name], err = conveyer.New(name, configConveyer)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"companent": "Conveyer" + name,
				"function":  "New()",
			}).Error(err)
		}
		wg.Add(1)
		go conveyers[name].Start(ctx, &wg)
	}
	go Collector(ctx, db, collectChan)
	for name, conv := range conveyers {
		TargetList := config.GetTargetList(name)

		go func(ctx context.Context, conv *conveyer.Conveyer, list []string) {
			count := 0
			defer func() {
				logger.WithFields(logrus.Fields{
					"component": "Conveyer:" + conv.GetName(),
					"status":    "Done",
				}).Infof("Check %d links", count)
				conv.Close()
			}()
			logger.WithFields(logrus.Fields{
				"component": "Conveyer:" + conv.GetName(),
				"status":    "Received",
			}).Infof("%d links", len(TargetList))
			for _, url := range list {
				conv.GetInput() <- []byte(url)
				count++
				select {
				case data := <-conv.GetOutput():
					strData := hex.EncodeToString(data)
					collectChan <- &model.Target{
						Url:  url,
						Hash: strData,
					}
				case msgErr := <-conv.GetError():
					logger.WithField("component", "conveyer:"+conv.GetName()).Error(msgErr)
				case <-ctx.Done():
					return
				}
			}
		}(ctx, conv, TargetList)

	}
	logger.Info("Wait all conveyer")
	wg.Wait()
	logger.Info("Finish factory")
}
