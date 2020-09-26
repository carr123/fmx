package fmx

import (
	"net/http"
	"strings"

	"github.com/CloudyKit/router"
)

//Router http router
type Router struct {
	handlers []HandlerFunc
	engine   *Engine
}

//Use register middleware
//these middlewares are combined and added to Router
func (r *Router) Use(middlewares ...HandlerFunc) *Router {
	r.handlers = append(r.handlers, middlewares...)
	return r
}

//GET handle GET method
func (r *Router) GET(path string, handlers ...HandlerFunc) {
	r.Handle("GET", path, handlers)
}

//POST handle POST method
func (r *Router) POST(path string, handlers ...HandlerFunc) {
	r.Handle("POST", path, handlers)
}

//PATCH handle PATCH method
func (r *Router) PATCH(path string, handlers ...HandlerFunc) {
	r.Handle("PATCH", path, handlers)
}

//PUT handle PUT method
func (r *Router) PUT(path string, handlers ...HandlerFunc) {
	r.Handle("PUT", path, handlers)
}

//DELETE handle DELETE method
func (r *Router) DELETE(path string, handlers ...HandlerFunc) {
	r.Handle("DELETE", path, handlers)
}

//HEAD handle HEAD method
func (r *Router) HEAD(path string, handlers ...HandlerFunc) {
	r.Handle("HEAD", path, handlers)
}

//OPTIONS handle OPTIONS method
func (r *Router) OPTIONS(path string, handlers ...HandlerFunc) {
	r.Handle("OPTIONS", path, handlers)
}

//TRACE handle TRACE method
func (r *Router) TRACE(path string, handlers ...HandlerFunc) {
	r.Handle("TRACE", path, handlers)
}

//CONNECT handle CONNECT method
func (r *Router) CONNECT(path string, handlers ...HandlerFunc) {
	r.Handle("CONNECT", path, handlers)
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE
func (r *Router) Any(path string, handlers ...HandlerFunc) {
	r.Handle("GET", path, handlers)
	r.Handle("POST", path, handlers)
	r.Handle("PUT", path, handlers)
	r.Handle("PATCH", path, handlers)
	r.Handle("HEAD", path, handlers)
	r.Handle("OPTIONS", path, handlers)
	r.Handle("DELETE", path, handlers)
	r.Handle("CONNECT", path, handlers)
	r.Handle("TRACE", path, handlers)
}

//Group group route
func (r *Router) Group(handlers ...HandlerFunc) *Router {
	handlers = r.combineHandlers(handlers)
	return &Router{
		handlers: handlers,
		engine:   r.engine,
	}
}

//Handle handle with specific method
func (r *Router) Handle(method, path string, handlers []HandlerFunc) {
	handlers = r.combineHandlers(handlers)

	r.engine.httprouter.AddRoute(method, path, func(w http.ResponseWriter, req *http.Request, params router.Parameter) {
		c := r.engine.createContext(w, req)
		c.params = &params
		c.handlers = handlers
		c.Next()
		c.Keys = nil
		c._errs = c._errs[:0]
		c._logs = c._logs[:0]
		r.engine.pool.Put(c)
	})
}

func (r *Router) ServeDir(prefix string, dir string) {
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	cutPrefixhandler := http.StripPrefix(prefix, FileServer2(http.Dir(dir)))
	h := func(c *Context) {
		cutPrefixhandler.ServeHTTP(c.Writer, c.Request)
	}

	handlers := r.combineHandlers([]HandlerFunc{HandlerFunc(h)})

	finalHandler := func(w http.ResponseWriter, req *http.Request, param router.Parameter) {
		c := r.engine.createContext(w, req)
		c.params = &param
		c.handlers = handlers
		c.Next()
		c.Keys = nil
		c._errs = c._errs[:0]
		c._logs = c._logs[:0]
		r.engine.pool.Put(c)
	}

	r.engine.httprouter.AddRoute("GET", prefix+"*filepath", finalHandler)
	r.engine.httprouter.AddRoute("HEAD", prefix+"*filepath", finalHandler)
	return
}

func (r *Router) combineHandlers(handlers []HandlerFunc) []HandlerFunc {
	h := make([]HandlerFunc, 0, len(r.handlers)+len(handlers))
	h = append(h, r.handlers...)

	for _, fn := range handlers {
		if fn != nil {
			h = append(h, fn)
		}
	}

	return h
}
