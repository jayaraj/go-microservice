package server

import (
	"context"
	"reflect"
	"sort"
)

type Descriptor struct {
	Name         string
	Instance     Service
	InitPriority Priority
}

type Service interface {
	Init() error
	OnConfig()
}

type BackgroundService interface {
	Run(ctx context.Context) error
}

type Priority int

const (
	High      Priority = 100
	InterHigh Priority = 75
	Inter     Priority = 50
	InterLow  Priority = 25
	Low       Priority = 0
)

var services []*Descriptor

func RegisterService(instance Service, priority Priority) {
	services = append(services, &Descriptor{
		Name:         reflect.TypeOf(instance).Elem().Name(),
		Instance:     instance,
		InitPriority: priority,
	})
}

func Register(descriptor *Descriptor) {
	services = append(services, descriptor)
}

func GetServices() []*Descriptor {
	sort.Slice(services, func(i, j int) bool {
		return services[i].InitPriority > services[j].InitPriority
	})
	return services
}
