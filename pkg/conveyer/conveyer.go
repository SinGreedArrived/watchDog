package conveyer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

// Conveyer { }
type Conveyer struct {
	name        string
	input       chan []byte
	output      chan []byte
	err         chan error
	data        [][]byte
	stack       [][]byte
	pointer     int
	httpClient  *http.Client
	httpRequest *http.Request
	steps       []func(*Conveyer) error
	config      *Config
}

// init conveyer by name name
func (self *Conveyer) init(name string) error {
	self.name = name
	self.pointer = 0
	self.data = make([][]byte, 0)
	self.data = append(self.data, []byte(""))
	self.stack = make([][]byte, 0)
	self.stack = append(self.data, []byte(""))
	self.input = make(chan []byte)
	self.output = make(chan []byte)
	self.err = make(chan error)
	self.httpClient = &http.Client{}
	httpTransport := &http.Transport{}
	self.httpRequest, _ = http.NewRequest("", "", nil)
	self.steps = make([]func(*Conveyer) error, 0)
	self.httpRequest.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:81.0) Gecko/20100101 Firefox/81.0")
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
	for _, command := range self.config.Steps {
		cmd := command
		err := self.AddStep(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

// Start conveyer
//	Example:
//		var wg sync.WaitGroup
//		...
//		wg.Add(1)
//		// Must using gorontine for Start()
//		go conv.Start(&wg)
//		...
func (self *Conveyer) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			self.err <- errors.New(fmt.Sprint(r))
			wg.Add(1)
			go self.Start(ctx, wg)
		}
		wg.Done()
	}()
	for elem := range self.input {
		self.WriteData(elem)
		for _, step := range self.steps {
			err := step(self)
			if err != nil {
				panic(err)
			}
		}
		select {
		case self.output <- self.GetData():
		case <-ctx.Done():
			return
		}
		time.Sleep(time.Millisecond * time.Duration(self.config.Delay))
	}
}

// Get data from buffer by index step. Default get last data
func (self *Conveyer) GetData(number ...int) []byte {
	if len(number) > 0 {
		return self.data[number[0]]
	}
	return self.data[self.pointer]
}

// Write data to buffer
func (self *Conveyer) WriteData(data []byte) {
	if self.pointer == len(self.data)-1 {
		self.data = append(self.data, data)
		self.pointer++
	} else {
		self.data[self.pointer] = data
	}
}

/*
Create new conveyer by config and name.
	Example:
		conv := conveyer.New("ifconfig", nil)
			OR
		conf := conveyer.NewConfig()
		conv := conveyer.New("ifconfig", conf)
*/
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

// Get input channel type []byte for write data to conveyer
func (self *Conveyer) GetInput() chan []byte {
	return self.input
}

// Get output channel type []byte for read data from conveyer
func (self *Conveyer) GetOutput() chan []byte {
	return self.output
}

// Get error channel type error for read error from conveyer
func (self *Conveyer) GetError() chan error {
	return self.err
}

// Get conveyer name
func (self *Conveyer) GetName() string {
	return self.name
}

// Close() close conveyer
func (self *Conveyer) Close() error {
	close(self.input)
	return nil
}
