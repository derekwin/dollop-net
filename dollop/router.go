package dollop

type RawRouterI interface {
	PreHandler(req RawRequestI)
	Handler(req RawRequestI)
	AfterHandler(req RawRequestI)
}

// inherit this BaseRouter while implement a new Router
type BaseRawRouter struct {
}

func (br BaseRawRouter) PreHandler(req RawRequestI) {}

func (br BaseRawRouter) Handler(req RawRequestI) {}

func (br BaseRawRouter) AfterHandler(req RawRequestI) {}

type FrameRouterI interface {
	PreHandler(req FrameRequestI)
	Handler(req FrameRequestI)
	AfterHandler(req FrameRequestI)
}

// inherit this BaseRouter while implement a new Router
type BaseFrameRouter struct {
}

func (br BaseFrameRouter) PreHandler(req FrameRequestI) {}

func (br BaseFrameRouter) Handler(req FrameRequestI) {}

func (br BaseFrameRouter) AfterHandler(req FrameRequestI) {}
