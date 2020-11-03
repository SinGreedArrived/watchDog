package conveyer

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/proxy"
)

func Panic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

type Config struct {
	Cookies map[string]string
	Proxy   string
	Timeout int `toml:"timeout"`
	Delay   int `toml:"delay"`
	Steps   [][]string
}

func NewConfig() *Config {
	return &Config{}
}

type conveyer struct {
	name        string
	input       chan string
	output      chan []byte
	data        [][]byte
	pointer     int
	httpClient  *http.Client
	httpRequest *http.Request
	steps       []func(*conveyer) error
	config      *Config
}

func (self *conveyer) init(name string, cc *Config) error {
	self.config = cc
	self.name = name
	self.pointer = 0
	self.input = make(chan string)
	self.output = make(chan []byte)
	self.httpClient = &http.Client{}
	httpTransport := &http.Transport{}
	self.httpRequest, _ = http.NewRequest("", "", nil)
	self.steps = make([]func(*conveyer) error, 0)
	delay := 0
	self.httpRequest.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:81.0) Gecko/20100101 Firefox/81.0")
	if cc.Proxy != "" {
		dialer, err := proxy.SOCKS5("tcp", cc.Proxy, nil, proxy.Direct)
		if err != nil {
			return err
		}
		httpTransport.Dial = dialer.Dial
	}
	if cc.Timeout != 0 {
		self.httpClient.Timeout = time.Millisecond * time.Duration(cc.Timeout)
	}
	if cc.Cookies != nil {
		for key, value := range cc.Cookies {
			self.httpRequest.AddCookie(&http.Cookie{Name: key, Value: value})
		}
	}
	delay++
	delay--
	self.httpClient.Transport = httpTransport
	self.data = make([][]byte, 0)
	self.data = append(self.data, []byte(self.httpRequest.URL.String()))
	for _, command := range self.config.Steps {
		cmd := command
		switch cmd[0] {
		case "download":
			self.steps = append(self.steps, func(conveyer *conveyer) error {
				elem, err := url.Parse(string(conveyer.GetData()))
				if err != nil {
					return err
				}
				conveyer.httpRequest.URL = elem
				resp, err := conveyer.httpClient.Do(conveyer.httpRequest)
				if err != nil {
					return err
				}
				if resp.StatusCode != 200 {
					return errors.New(fmt.Sprintf("Status code: %d", resp.StatusCode))
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				conveyer.WriteData(body)
				return nil
			})
		case "remove":
			self.steps = append(self.steps, func(conveyer *conveyer) error {
				reg := cmd[1]
				regxp, err := regexp.Compile(reg)
				if err != nil {
					return err
				}
				data := regxp.ReplaceAll(conveyer.data[conveyer.pointer], []byte(""))
				conveyer.WriteData(data)
				return nil
			})
		case "replace":
			self.steps = append(self.steps, func(conveyer *conveyer) error {
				reg := cmd[1]
				repl := cmd[2]
				regxp, err := regexp.Compile(reg)
				if err != nil {
					return err
				}
				data := regxp.ReplaceAll(conveyer.data[conveyer.pointer], []byte(repl))
				conveyer.WriteData(data)
				return nil
			})
		case "css":
			self.steps = append(self.steps, func(conveyer *conveyer) error {
				selector := cmd[1]
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(conveyer.data[conveyer.pointer])))
				if err != nil {
					return err
				}
				slc := doc.Find(selector)
				body, err := goquery.OuterHtml(slc)
				if err != nil {
					return err
				}
				conveyer.WriteData([]byte(body))
				return nil
			})
		case "undo":
			self.steps = append(self.steps, func(conveyer *conveyer) error {
				i, _ := strconv.Atoi(cmd[1])
				for ; i != 0; i-- {
					conveyer.pointer--
					if conveyer.pointer < 0 {
						conveyer.pointer = 0
					}
				}
				return nil
			})
		case "redo":
			self.steps = append(self.steps, func(conveyer *conveyer) error {
				i, _ := strconv.Atoi(cmd[1])
				for ; i != 0; i-- {
					conveyer.pointer++
					if conveyer.pointer == len(conveyer.data) {
						conveyer.pointer--
					}
				}
				return nil
			})
		case "md5":
			self.steps = append(self.steps, func(conveyer *conveyer) error {
				hash := md5.Sum(conveyer.data[conveyer.pointer])
				conveyer.WriteData([]byte(hash[:]))
				return nil
			})
		case "show":
			self.steps = append(self.steps, func(conveyer *conveyer) error {
				fmt.Println(string(conveyer.data[conveyer.pointer]))
				return nil
			})
		}
	}
	return nil
}

func (self *conveyer) Start(wg *sync.WaitGroup) chan string {
	wg.Add(1)
	go func() {
		for elem := range self.input {
			self.WriteData([]byte(elem))
			for _, step := range self.steps {
				err := step(self)
				if err != nil {
					log.Panic(err)
				}
			}
			self.output <- self.GetData()
		}
		defer func() {
			if r := recover(); r != nil {
				wg.Add(1)
				self.Start(wg)
			}
			wg.Done()
		}()
	}()
	return self.input
}

func (self *conveyer) GetData(number ...int) []byte {
	if len(number) > 0 {
		return self.data[number[0]]
	}
	return self.data[self.pointer]
}

func (self *conveyer) WriteData(data []byte) {
	if self.pointer == len(self.data)-1 {
		self.data = append(self.data, data)
		self.pointer++
	} else {
		self.data[self.pointer] = data
	}
}

func New(name string, cc *Config) *conveyer {
	tmp := &conveyer{}
	tmp.init(name, cc)
	return tmp
}

func (self *conveyer) GetInputChan() chan string {
	return self.input
}

func (self *conveyer) GetOutputChan() chan []byte {
	return self.output
}