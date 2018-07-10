package fmx

import (
	"net/http"
	"sync"

	"github.com/CloudyKit/router"
)

type Engine struct {
	*Router
	httprouter *router.Router
	pool       sync.Pool
}

func New() *Engine {
	engine := &Engine{}
	engine.Router = &Router{
		handlers: nil,
		engine:   engine,
	}

	engine.httprouter = router.New()
	engine.pool.New = func() interface{} {
		c := &Context{}
		c.Writer = &writerImpl{}
		c.index = -1
		return c
	}

	return engine
}

// Conforms to the http.Handler interface.
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	engine.httprouter.ServeHTTP(w, req)
}
