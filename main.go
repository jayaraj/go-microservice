package main

import (
	"flag"
	_ "go-microservice/infra/cache"
	_ "go-microservice/infra/dbs/postgres"
	_ "go-microservice/infra/gateway"
	"go-microservice/infra/server"
	_ "go-microservice/repository"
	_ "go-microservice/services"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	homepath := flag.String("homepath", ".", "path to service install/home path, defaults to working directory")
	configpath := flag.String("configpath", *homepath+"/conf/default.yml", "path to configfile, defaults to working directory/conf/default.yml")
	flag.Parse()

	server := server.NewServer(*homepath, *configpath)
	go listenToSystemSignals(server)

	err := server.Run()

	code := server.ExitCode(err)
	os.Exit(code)
}

func listenToSystemSignals(server *server.Server) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case sig := <-signalChan:
			reason := "System signal: " + sig.String()
			server.Shutdown(reason)
		}
	}
}
