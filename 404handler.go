package fmx

import (
	"net/http"
)

type notfound struct {
	Realwriter http.ResponseWriter
	Is404      bool
}

func (c *notfound) Header() http.Header {
	return c.Realwriter.Header()
}

func (c *notfound) Write(p []byte) (int, error) {
	if c.Is404 {
		return len(p), nil
	}

	return c.Realwriter.Write(p)
}

func (c *notfound) WriteHeader(status int) {
	if status == 404 {
		c.Is404 = true
		return
	}

	c.Realwriter.WriteHeader(status)
}

func Code404Handler(fn HandlerFunc) HandlerFunc {
	return func(c *Context) {
		originWriter := c.Writer
		h := &notfound{Realwriter: c.Writer, Is404: false}
		c.Writer = NewWriter(h)
		c.Next()
		c.Writer = originWriter

		if h.Is404 && fn != nil {
			fn(c)
		}
	}
}
