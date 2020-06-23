# go-microservice 
![Microservice](https://github.com/jayaraj/go-microservice/workflows/Microservice/badge.svg)

This is a template for writing a micro-service in go with a server infrastructure. It uses [`grpc-gateway`](https://github.com/grpc-ecosystem/grpc-gateway) for endpoints which supports both GRPC & REST interfaces.

All the packages are loosely coupled so just remove them if not required. Make sure to clean imported infra packages from main.go if not used.

## Application Components
1. `proto`
    - Proto definitions for your service. (e.g.)user.proto file defines your service interfaces.
    - This service supports swagger UI. Make sure you change the `yourservice.swagger.json` within in `proto/openapi/index.html`
2. `dtos` 
   - Repository models, command structs to communicate between components.
3. `services`
   - Application files to service the request
4. `repository`
    - Optional if you access database
5. `conf`
    - default.yml which used by default. main.go supports custom config file path.

Every component file within this service will have four functions

1. `init` register your component structure with server to initialize the component with priority or to start your background services.
```go
func init() {
	server.RegisterService(&userRepo{}, server.Low)
}
```
2. `Init` function where you intialize your component, register your services with bus for serving other components
```go
func (c *userRepo) Init() (err error) {

	c.addUserMigrations()

	//Register for all the repository requests
	bus.AddHandler(CreateUser)
	bus.AddHandler(ListUsers)
	return nil
}
```

3. `Run` (optional)background service. `<-ctx.Done()` will return err when server recieves termination, use this for safe shutdown.
```go
func (c *GRPC) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		log.Infoln("Stopping Grpc")
		c.grpcServer.GracefulStop()
	}()
	log.WithField("Port", c.grpcPort).Info("Grpc listening...")
	return c.grpcServer.Serve(c.listener)
}
```

4. `OnConfig` called when config file is edited
```go
func (service *UserService) OnConfig() {
}
```

## Infra
1. `bus`
    - Use bus to communicate between components, avoid circular imports 
2. `cache`
    - Three cache libraries are supported. Use the ones you need and remove others.
3. `db`
    - Supports postgres incremental migration with [`gorm`](https://gorm.io/) 
4. `gateway`
    - [`grpc-gateway`](https://github.com/grpc-ecosystem/grpc-gateway) wrappers. 

## Installation

1. Run
   - `make setup` installs all the modules required
   - `make generate` compiles the proto files
   - `go run .`
   
2. Deploy in `docker`
   - `docker-compose up -d`


