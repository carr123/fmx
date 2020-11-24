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
