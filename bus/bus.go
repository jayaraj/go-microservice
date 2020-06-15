package bus

import (
	"context"
	"errors"
	"reflect"
)

var (
	instance          Bus
	ErrMissingHandler = errors.New("Handler Not Found")
)

type HandlerFunc interface{}

type CtxHandlerFunc func()

type Msg interface{}

type Bus interface {
	dispatch(msg Msg) error
	dispatchCtx(ctx context.Context, msg Msg) error
	publish(msg Msg) error
	addHandler(handler HandlerFunc)
	addHandlerCtx(handler HandlerFunc)
	addEventListener(handler HandlerFunc)
}

type bus struct {
	handlers        map[string]HandlerFunc
	handlersWithCtx map[string]HandlerFunc
	listeners       map[string][]HandlerFunc
}

func Dispatch(msg Msg) error {
	return instance.dispatch(msg)
}

func DispatchCtx(ctx context.Context, msg Msg) error {
	return instance.dispatchCtx(ctx, msg)
}

func Publish(msg Msg) error {
	return instance.publish(msg)
}

func AddHandler(handler HandlerFunc) {
	instance.addHandler(handler)
}

func AddHandlerCtx(handler HandlerFunc) {
	instance.addHandlerCtx(handler)
}

func AddEventListener(handler HandlerFunc) {
	instance.addEventListener(handler)
}

func init() {
	instance = &bus{
		handlers:        make(map[string]HandlerFunc),
		handlersWithCtx: make(map[string]HandlerFunc),
		listeners:       make(map[string][]HandlerFunc),
	}
}

func (bus *bus) dispatchCtx(ctx context.Context, msg Msg) error {
	var msgName = reflect.TypeOf(msg).Elem().Name()

	var handler = bus.handlersWithCtx[msgName]
	if handler == nil {
		return ErrMissingHandler
	}

	var params = []reflect.Value{}
	params = append(params, reflect.ValueOf(ctx))
	params = append(params, reflect.ValueOf(msg))

	ret := reflect.ValueOf(handler).Call(params)
	err := ret[0].Interface()
	if err == nil {
		return nil
	}
	return err.(error)
}

func (bus *bus) dispatch(msg Msg) error {
	var msgName = reflect.TypeOf(msg).Elem().Name()

	var handler = bus.handlersWithCtx[msgName]
	withCtx := true

	if handler == nil {
		withCtx = false
		handler = bus.handlers[msgName]
	}

	if handler == nil {
		return ErrMissingHandler
	}

	var params = []reflect.Value{}
	if withCtx {
		params = append(params, reflect.ValueOf(context.Background()))
	}
	params = append(params, reflect.ValueOf(msg))

	ret := reflect.ValueOf(handler).Call(params)
	err := ret[0].Interface()
	if err == nil {
		return nil
	}
	return err.(error)
}

func (bus *bus) publish(msg Msg) error {
	var msgName = reflect.TypeOf(msg).Elem().Name()
	var listeners = bus.listeners[msgName]

	var params = make([]reflect.Value, 1)
	params[0] = reflect.ValueOf(msg)

	for _, listenerHandler := range listeners {
		ret := reflect.ValueOf(listenerHandler).Call(params)
		err := ret[0].Interface()
		if err != nil {
			return err.(error)
		}
	}

	return nil
}

func (bus *bus) addHandler(handler HandlerFunc) {
	handlerType := reflect.TypeOf(handler)
	queryTypeName := handlerType.In(0).Elem().Name()
	bus.handlers[queryTypeName] = handler
}

func (bus *bus) addHandlerCtx(handler HandlerFunc) {
	handlerType := reflect.TypeOf(handler)
	queryTypeName := handlerType.In(1).Elem().Name()
	bus.handlersWithCtx[queryTypeName] = handler
}

func (bus *bus) addEventListener(handler HandlerFunc) {
	handlerType := reflect.TypeOf(handler)
	eventName := handlerType.In(0).Elem().Name()
	_, exists := bus.listeners[eventName]
	if !exists {
		bus.listeners[eventName] = make([]HandlerFunc, 0)
	}
	bus.listeners[eventName] = append(bus.listeners[eventName], handler)
}
