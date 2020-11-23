package main

import (
	"flag"
	"net"
	"strconv"
	"projects/parser/pkg/api"
	"projects/parser/pkg/factory"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	port   			int
	logger      *logrus.Logger
)

func init() {
	flag.IntVar(&port, "port", 8855, "port for server")
	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		HideKeys:        true,
	})
}

func main() {
	logger.Info("Start factory server on port:"+strconv.Itoa(port))
	flag.Parse()

	server := grpc.NewServer()
	service := &factory.GRPCServer{}
	api.RegisterFactoryServer(server, service)

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}	
	defer ln.Close()
	if err := server.Serve(ln); err != nil {
		panic(err)
	}
}
