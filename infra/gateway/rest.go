package gateway

import (
	"context"
	"fmt"
	"go-microservice/infra/server"
	"mime"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rakyll/statik/fs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	_ "go-microservice/statik"
)

var restInstance *REST

type REST struct {
	httpPort   int
	grpcPort   int
	connection *grpc.ClientConn
	mux        *runtime.ServeMux
}

func init() {
	restInstance = &REST{}
	//Set it High or low based on your requirement
	server.RegisterService(restInstance, server.Low)
}

func (c *REST) Init() (err error) {
	viper.SetDefault("http", 9000)
	viper.SetDefault("grpc", 9001)
	c.grpcPort = viper.GetInt("grpc")
	c.httpPort = viper.GetInt("http")
	address := fmt.Sprintf("dns:///0.0.0.0:%d", c.grpcPort)
	c.connection, err = grpc.DialContext(context.Background(), address, grpc.WithInsecure())
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("OpenAPI failed to dial Grpc")
		return err
	}
	c.mux = runtime.NewServeMux()
	return err
}

func (c *REST) OnConfig() {
}

func (c *REST) Run(ctx context.Context) error {
	handler, err := openAPIHandler()
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("OpenAPI creation failed")
		return err
	}
	address := fmt.Sprintf("0.0.0.0:%d", c.httpPort)
	server := &http.Server{
		Addr: address,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				c.mux.ServeHTTP(w, r)
				return
			}
			handler.ServeHTTP(w, r)
		}),
	}
	go func() {
		<-ctx.Done()
		log.Info("Stopping OpenAPI")
		server.Shutdown(ctx)
	}()
	log.WithField("Port", c.httpPort).Info("OpenAPI listening...")
	return server.ListenAndServe()
}

func openAPIHandler() (http.Handler, error) {
	mime.AddExtensionType(".svg", "image/svg+xml")

	statikFS, err := fs.New()
	if err != nil {
		return nil, err
	}
	return http.FileServer(statikFS), nil
}
