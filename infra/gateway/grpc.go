package gateway

import (
	"context"
	"fmt"
	"go-microservice/infra/server"
	"net"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var rpcInstance *GRPC

type GRPC struct {
	grpcPort   int
	listener   net.Listener
	grpcServer *grpc.Server
}

func init() {
	rpcInstance = &GRPC{}
	server.RegisterService(rpcInstance, server.High)
}

func (c *GRPC) Init() (err error) {

	viper.SetDefault("grpc", 9001)
	c.grpcPort = viper.GetInt("grpc")
	address := fmt.Sprintf("0.0.0.0:%d", c.grpcPort)
	c.listener, err = net.Listen("tcp", address)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
			"Port":  c.grpcPort,
		}).Error("Grpc server failed to listen")
		return err
	}
	c.grpcServer = grpc.NewServer()
	return nil
}

func (c *GRPC) OnConfig() {
}

func (c *GRPC) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		log.Infoln("Stopping Grpc")
		c.grpcServer.GracefulStop()
	}()
	log.WithField("Port", c.grpcPort).Info("Grpc listening...")
	return c.grpcServer.Serve(c.listener)
}
