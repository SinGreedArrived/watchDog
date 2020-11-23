package factory

import (
	"context"
	"encoding/hex"
	"errors"
	"projects/parser/internal/conveyer"
	"projects/parser/pkg/api"
	"sync"

	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
	nested "github.com/antonfisher/nested-logrus-formatter"
)

type GRPCServer struct {
	wg sync.WaitGroup
	convList map[string]*conveyer.Conveyer
	api.UnimplementedFactoryServer
	logger *logrus.Logger
}

func (self *GRPCServer) Init() {
	self.convList = make(map[string]*conveyer.Conveyer)
	self.logger = logrus.New()
	self.logger.SetLevel(logrus.InfoLevel)
	self.logger.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		HideKeys:        true,
	})
}

func (self *GRPCServer) CreateConveyer(_ context.Context, req *api.Request) (*api.Response, error) {
	self.logger.Infof("Create conveyer name: %s | config: %s",req.GetName(), req.GetUrl())
	if self.convList == nil {
		self.Init()
	}
	config := conveyer.NewConfig()
	err := toml.Unmarshal(req.GetUrl(), config)
	if err != nil {
		return nil, err
	}
	conv,err := conveyer.New(req.GetName(), config)
	if err != nil {
		return nil, err
	}
	go conv.Start(context.Background(), &self.wg)
	self.convList[req.GetName()] = conv
	return &api.Response{Result:"Ok"},nil
}

func (self *GRPCServer) Do(_ context.Context, req *api.Request) (*api.Response, error) {
	if _, ok := self.convList[req.GetName()]; ok{
			self.convList[req.GetName()].GetInput() <- req.Url
		select {
		case msg := <- self.convList[req.GetName()].GetOutput():
				dst := make([]byte, hex.EncodedLen(len(msg)))
				hex.Encode(dst, msg)
			return &api.Response{Result:string(dst)}, nil
		case err := <- self.convList[req.GetName()].GetError():
			return nil, err
		}
	}
	return nil, errors.New("I don't have conveyer by name:" + req.GetName())
}
