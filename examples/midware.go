package main

import (
	"github.com/carr123/fmx"
)

//http basic auth
func BasicAuth() func(c *fmx.Context) {
	return func(c *fmx.Context) {
		var bCheckAuth bool = false
		username, password, ok := c.Request.BasicAuth()
		if ok && username == "root" && password == "1234" {
			bCheckAuth = true
		}

		if !bCheckAuth {
			c.Writer.Header().Add("WWW-Authenticate", `Basic realm=""`)
			c.String(401, "Unauthorized")
			c.Abort()
			return
		}

		c.Set("username", username)
		c.Next()
	}
}

//cross origin
func XOrigin() func(c *fmx.Context) {
	return func(c *fmx.Context) {
		c.Next()
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	}
}
