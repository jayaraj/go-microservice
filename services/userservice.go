package services

import (
	"context"
	"go-microservice/dtos"
	gw "go-microservice/generated/gateway/proto"
	"go-microservice/generated/proto"
	"go-microservice/infra/bus"
	"go-microservice/infra/gateway"
	"go-microservice/infra/server"
	"sync"

	log "github.com/sirupsen/logrus"
)

type UserService struct {
	mu *sync.RWMutex
}

func init() {
	server.RegisterService(&UserService{
		mu: &sync.RWMutex{},
	}, server.Low)
}

func (service *UserService) Init() (err error) {

	//Register this service with Grpc alone with service description generated
	// for this service at <service>.pb.go
	proto.RegisterUserServiceServer(gateway.Grpc(), service)

	//Call RegisterServiceHandler generated at <service>.pb.gw.go
	gw.RegisterUserServiceHandler(context.Background(), gateway.Mux(), gateway.ClientConnection())
	return nil
}

func (service *UserService) OnConfig() {
}

func (service *UserService) AddUser(ctx context.Context, request *proto.AddUserRequest) (*proto.AddUserResponse, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	cmd := dtos.CreateUserCmd{
		Name:  request.GetName(),
		Email: request.GetEmail(),
	}
	if err := bus.Dispatch(&cmd); err != nil {
		log.WithField("Error", err).Error("Add user failed")
		return nil, err
	}
	user := proto.AddUserResponse{
		Id:    cmd.Result.Id,
		Name:  cmd.Result.Name,
		Email: cmd.Result.Email,
	}
	return &user, nil
}

func (service *UserService) ListUsers(request *proto.ListUsersRequest, srv proto.UserService_ListUsersServer) error {
	service.mu.RLock()
	defer service.mu.RUnlock()

	cmd := dtos.ListUsersCmd{
		Limit: request.GetLimit(),
		Page:  request.GetPage(),
	}
	if err := bus.Dispatch(&cmd); err != nil {
		log.WithField("Error", err).Error("List users failed")
		return err
	}
	for _, user := range cmd.Result.Users {
		err := srv.Send(&proto.ListUsersResponse{
			Id:    user.Id,
			Name:  user.Name,
			Email: user.Email,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
