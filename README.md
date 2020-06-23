# go-microservice 
![Microservice](https://github.com/jayaraj/go-microservice/workflows/Microservice/badge.svg)

This is a template for writing a micro-service in go with a server infrastructure. It uses [`grpc-gateway`](https://github.com/grpc-ecosystem/grpc-gateway) for endpoints which supports both GRPC & REST interfaces.

All the packages are loosely coupled so just remove them if not required. Make sure to clean imported infra packages from main.go if not used.

## Application Components
1. `proto`
    - Proto definitions for your service.
2. `dtos` 
   - Repository models, command structs to communicate between components.
3. `services`
   - Application files to service the request
4. `repository`
    - Optional if you access database
5. `conf`
    - default.yml which used by default. main.go supports custom config file path.

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


