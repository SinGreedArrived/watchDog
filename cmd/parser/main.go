package main

import (
	"flag"
	"fmt"
	"log"
	"projects/parser/internal/app/configure"
	"projects/parser/internal/app/conveyer"
	"sync"
)

var (
	CONFIG_PATH string
	wg          sync.WaitGroup
)

func init() {
	flag.StringVar(&CONFIG_PATH, "config", "configs/config.toml", "path to config file")
}

func main() {
	flag.Parse()
	config := configure.New()
	err := config.LoadToml(CONFIG_PATH)
	if err != nil {
		log.Fatal(err)
	}
	conv := conveyer.New("remanga.org", config.GetConveyerConfig("remanga.org"))
	conv.Start(&wg)
	conv.GetInputChan() <- "https://remanga.org/manga/the_beginning_after_the_end?subpath=content"
	fmt.Println(<-conv.GetOutputChan())
}
