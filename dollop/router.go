package dollop

import "fmt"

type RawRouterI interface {
	PreHandler(req RawRequestI) error
	Handler(req RawRequestI) error
	AfterHandler(req RawRequestI) error
}

// inherit this BaseRouter while implement a new Router
type BaseRawRouter struct {
}

func (br BaseRawRouter) PreHandler(req RawRequestI) error {
	return nil
}

func (br BaseRawRouter) Handler(req RawRequestI) error {
	fmt.Println("default raw router preHandler")
	return nil
}

func (br BaseRawRouter) AfterHandler(req RawRequestI) error {
	return nil
}

type FrameRouterI interface {
	PreHandler(req FrameRequestI) error
	Handler(req FrameRequestI) error
	AfterHandler(req FrameRequestI) error
}

// inherit this BaseRouter while implement a new Router
type BaseFrameRouter struct {
}

func (br BaseFrameRouter) PreHandler(req FrameRequestI) error {
	return nil
}

func (br BaseFrameRouter) Handler(req FrameRequestI) error {
	fmt.Println("default frame router preHandler")
	return nil
}

func (br BaseFrameRouter) AfterHandler(req FrameRequestI) error {
	return nil
}
