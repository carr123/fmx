package fmx

import (
	"sync"

	"github.com/CloudyKit/router"
)

type Engine struct {
	*Router
	httprouter *router.Router
	pool       sync.Pool
}

func NewServeMux() *Engine {
	engine := &Engine{}
	engine.Router = &Router{
		handlers: nil,
		engine:   engine,
	}

	engine.httprouter = router.New()
	engine.pool.New = func() interface{} {
		c := &Context{}
		c.Writer = &WriterImpl{}
		c.index = -1
		return c
	}

	return engine
}
