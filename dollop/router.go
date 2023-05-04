package dollop

type RouterI interface {
	PreHandler(req RequestI)
	Handler(req RequestI)
	AfterHandler(req RequestI)
}

// inherit this BaseRouter while implement a new Router
type BaseRouter struct {
}

func (br BaseRouter) PreHandler(req RequestI) {}

func (br BaseRouter) Handler(req RequestI) {}

func (br BaseRouter) AfterHandler(req RequestI) {}
