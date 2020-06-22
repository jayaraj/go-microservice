package gateway

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func Grpc() *grpc.Server {
	return rpcInstance.grpcServer
}

func ClientConnection() *grpc.ClientConn {
	return restInstance.connection
}

func Mux() *runtime.ServeMux {
	return restInstance.mux
}
