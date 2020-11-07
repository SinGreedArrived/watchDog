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

type Conveyer struct {
	name        string
	input       chan []byte
	output      chan []byte
	data        [][]byte
	pointer     int
	httpClient  *http.Client
	httpRequest *http.Request
	steps       []func(*Conveyer) error
	config      *Config
	//	logger      *logrus.Logger
}

func (self *Conveyer) init(name string) error {
	self.name = name
	self.pointer = 0
	self.input = make(chan []byte)
	self.output = make(chan []byte)
	self.httpClient = &http.Client{}
	httpTransport := &http.Transport{}
	self.httpRequest, _ = http.NewRequest("", "", nil)
	self.steps = make([]func(*Conveyer) error, 0)
	self.httpRequest.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:81.0) Gecko/20100101 Firefox/81.0")
	/*
		self.logger = logrus.New()
		if level, err := logrus.ParseLevel(self.config.logLevel); err != nil {
			self.logger.SetLevel(logrus.InfoLevel)
		} else {
			self.logger.SetLevel(level)
		}
	*/
	if self.config.Proxy != "" {
		dialer, err := proxy.SOCKS5("tcp", self.config.Proxy, nil, proxy.Direct)
		if err != nil {
			return err
		}
		httpTransport.Dial = dialer.Dial
	}
	if self.config.Timeout != 0 {
		self.httpClient.Timeout = time.Millisecond * time.Duration(self.config.Timeout)
	}
	if self.config.Cookies != nil {
		for key, value := range self.config.Cookies {
			self.httpRequest.AddCookie(&http.Cookie{Name: key, Value: value})
		}
	}
	self.httpClient.Transport = httpTransport
	self.data = make([][]byte, 0)
	self.data = append(self.data, []byte(self.httpRequest.URL.String()))
	for _, command := range self.config.Steps {
		cmd := command
		switch cmd[0] {
		case "download":
			self.steps = append(self.steps, func(conveyer *Conveyer) error {
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
			self.steps = append(self.steps, func(conveyer *Conveyer) error {
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
			self.steps = append(self.steps, func(conveyer *Conveyer) error {
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
			self.steps = append(self.steps, func(conveyer *Conveyer) error {
				selector := cmd[1]
				number := 0
				if len(cmd) > 2 {
					number, _ = strconv.Atoi(cmd[2])
				}
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(conveyer.data[conveyer.pointer])))
				if err != nil {
					return err
				}
				slc := doc.Find(selector).Eq(number)
				body, err := goquery.OuterHtml(slc)
				if err != nil {
					return err
				}
				conveyer.WriteData([]byte(body))
				return nil
			})
		case "undo":
			self.steps = append(self.steps, func(conveyer *Conveyer) error {
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
			self.steps = append(self.steps, func(conveyer *Conveyer) error {
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
			self.steps = append(self.steps, func(conveyer *Conveyer) error {
				hash := md5.Sum(conveyer.data[conveyer.pointer])
				conveyer.WriteData([]byte(hash[:]))
				return nil
			})
		case "show":
			self.steps = append(self.steps, func(conveyer *Conveyer) error {
				fmt.Println(string(conveyer.data[conveyer.pointer]))
				return nil
			})
		default:
			return errors.New(fmt.Sprintf("Command %s unknown", cmd[0]))
		}
	}
	return nil
}

func (self *Conveyer) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		for elem := range self.input {
			self.WriteData(elem)
			for _, step := range self.steps {
				err := step(self)
				if err != nil {
					Panic(err)
				}
			}
			self.output <- self.GetData()
			time.Sleep(time.Millisecond * time.Duration(self.config.Delay))
		}
		defer func() {
			if r := recover(); r != nil {
				self.Start(wg)
			}
			close(self.output)
			wg.Done()
		}()
	}()
}

func (self *Conveyer) SetOutput(output chan []byte) {
	self.output = output
}

func (self *Conveyer) GetData(number ...int) []byte {
	if len(number) > 0 {
		return self.data[number[0]]
	}
	return self.data[self.pointer]
}

func (self *Conveyer) WriteData(data []byte) {
	if self.pointer == len(self.data)-1 {
		self.data = append(self.data, data)
		self.pointer++
	} else {
		self.data[self.pointer] = data
	}
}

func New(name string, cc *Config) (*Conveyer, error) {
	if cc == nil {
		cc = NewConfig()
	}
	tmp := &Conveyer{
		config: cc,
	}
	err := tmp.init(name)
	if rerr := recover(); rerr != nil {
		return nil, errors.New(rerr.(string))
	}
	if err != nil {
		return nil, err
	}
	return tmp, nil
}

func (self *Conveyer) GetInput() chan []byte {
	return self.input
}

func (self *Conveyer) GetOutput() chan []byte {
	return self.output
}

func (self *Conveyer) GetName() string {
	return self.name
}

/*
func (self *Conveyer) SetLogger(logger *logrus.Logger) {
	self.logger = logger
}
*/

func (self *Conveyer) Close() {
	close(self.input)
}
