package fmx

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type LoggerFunc func(c *Context, szLog []byte)

func Logger() HandlerFunc {
	return LoggerWithFunc(DefaultLoggerFunc())
}

func LoggerWithFunc(fn LoggerFunc) HandlerFunc {
	if fn == nil {
		return func(c *Context) {

		}
	}

	return func(c *Context) {
		start := time.Now()
		logWriter := &bytes.Buffer{}

		szIP := c.ClientIP()
		rawPath := c.Request.URL.String()
		method := c.Request.Method
		szTimeBegin := start.Format("2006-01-02 15:04:05")

		defer func() {
			if e := recover(); e != nil {
				io.WriteString(logWriter, fmt.Sprintf("\r\n[mtx] %s | %s [500] %s %s\r\n", szTimeBegin, szIP, method, rawPath))
				stack := stack(3)
				panicInfo := fmt.Sprintf("PANIC: %s\r\n%s", e, stack)
				c.AddError(NewError(panicInfo))
				io.WriteString(logWriter, panicInfo)
				c.Writer.WriteHeader(http.StatusInternalServerError)
				fn(c, logWriter.Bytes())
			}
		}()

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		code := c.Writer.GetStatusCode()

		szLatency := fmt.Sprintf("%5.3fms", float64(latency)/float64(time.Millisecond))

		io.WriteString(logWriter, fmt.Sprintf("[mtx] %s |%3d| %s [%s] %s %s", szTimeBegin, code, szLatency, szIP, method, rawPath))
		for _, err := range c.GetErrors() {
			if errpos, ok := err.(ErrorWithPos); ok {
				io.WriteString(logWriter, "\r\n"+errpos.String()+"\r\n")
			} else {
				io.WriteString(logWriter, "\r\n"+err.Error())
			}
		}

		for _, szInfo := range c.GetLogs() {
			io.WriteString(logWriter, "\r\n"+szInfo)
		}

		io.WriteString(logWriter, "\r\n")
		fn(c, logWriter.Bytes())
	}
}

func DefaultLoggerFunc() LoggerFunc {
	return LoggerFunc(func(c *Context, szLog []byte) {
		io.Copy(os.Stdout, bytes.NewReader(szLog))
	})
}
