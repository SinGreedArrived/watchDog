package main

import (
	"context"
	//"encoding/hex"
	"flag"
	"os"
	"sync"

	"projects/parser/internal/configure"
	"projects/parser/pkg/api"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	configPath  string
	browserName string
	openNew     bool
	clearNew    bool
	notify      bool
	delayOpen   int
	addr        string
	logger      *logrus.Logger
	wg          sync.WaitGroup
)

func init() {
	dirName, _ := os.Getwd()
	flag.StringVar(&configPath, "config", dirName+"/configs/config.toml", "path to config file")
	flag.StringVar(&browserName, "browser", "firefox", "browser for open link")
	flag.StringVar(&addr, "addr", "127.0.0.1:8855", "address:port server")
	flag.BoolVar(&openNew, "open", false, "browser for open link")
	flag.BoolVar(&notify, "notify", false, "Notify for news")
	flag.BoolVar(&clearNew, "clear", false, "clear news")
	flag.IntVar(&delayOpen, "delay", 1500, "delay for open link")
	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		HideKeys:        true,
	})
}

func main() {
	logger.Info("Start factory...")
	flag.Parse()
	config := configure.NewConfig()
	err := config.LoadToml(configPath)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"package":  "configure",
			"function": "LoadToml",
		}).Fatal(err)
	}
	logger.WithField("component", "config").Info("Loaded configure from " + configPath)

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	cfg, err := config.GetConv("remanga.org")
	if err != nil {
		panic(err)
	}
	client := api.NewFactoryClient(conn)
	response,err := client.CreateConveyer(context.Background(), &api.Request{Name:"test", Url:cfg})
	if err != nil {
		panic(err)
	}
	logger.Info(response.GetResult())
	for _, url := range config.GetTargetList("remanga.org") {
		testTarget := []byte(url)
		response, err = client.Do(context.Background(), &api.Request{Name:"test", Url:testTarget})
		if err != nil {
			logger.Error(err)
			continue
		}
		logger.Info(response.GetResult())
	}
}
