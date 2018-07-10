package fmx

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"
)

func FullLogger() HandlerFunc {
	return FullLoggerWithFunc(DefaultLoggerFunc())
}

func _writeRequestLog(c *Context, t time.Time, logwriter io.Writer) {
	szIP := c.ClientIP()
	szProto := c.Request.Proto
	rawPath := c.Request.URL.String()
	method := c.Request.Method

	var reqHeader bytes.Buffer
	c.Request.Header.Write(&reqHeader)

	var reqBody string
	if c.Request.Body != nil {
		buff := new(bytes.Buffer)
		io.Copy(buff, c.Request.Body)
		c.Request.Body.Close()
		reqBody = buff.String()
		c.Request.Body = ioutil.NopCloser(buff)
	}

	io.WriteString(logwriter, "\r\n<----------------------LOG BEGIN----------------------------->\r\n")
	io.WriteString(logwriter, "requesttime:"+t.Format("2006-01-02 15:04:05")+"\r\n")
	io.WriteString(logwriter, "clientaddr:"+szIP+"\r\n\r\n")
	io.WriteString(logwriter, method+" "+rawPath+"  "+szProto+"\r\n")
	io.WriteString(logwriter, "Host:"+c.Request.Host+"\r\n")
	io.WriteString(logwriter, reqHeader.String()+"\r\n")

	if len(reqBody) > 0 {
		io.WriteString(logwriter, "\r\n"+reqBody+"\r\n")
	}
}

func _writeResponseLog(c *Context, logwriter io.Writer, code int, header http.Header, respBody *bytes.Buffer) {
	io.WriteString(logwriter, "----------------\r\n")
	io.WriteString(logwriter, "statuscode:"+fmt.Sprintf("%03d", code)+"\r\n")

	for k, v := range header {
		c.Writer.Header().Set(k, v[0])
	}

	c.Writer.WriteHeader(code)

	var respHeader bytes.Buffer
	c.Writer.Header().Write(&respHeader)
	io.WriteString(logwriter, respHeader.String()+"\r\n")

	if respBody.Len() > 0 {
		io.WriteString(logwriter, "\r\n"+respBody.String()+"\r\n")
		io.Copy(c.Writer, respBody)
	}
}

func FullLoggerWithFunc(fn LoggerFunc) HandlerFunc {
	if fn == nil {
		return func(c *Context) {

		}
	}

	return func(c *Context) {
		start := time.Now()

		logWriter := &bytes.Buffer{}
		_writeRequestLog(c, start, logWriter)

		rec := httptest.NewRecorder()
		originWriter := c.Writer
		c.Writer = NewWriter(rec)

		defer func() {
			if e := recover(); e != nil {
				io.WriteString(logWriter, "--------------------\r\n")
				stack := stack(3)
				panicInfo := fmt.Sprintf("PANIC: %s\r\n%s", e, stack)
				c.AddError(NewError(panicInfo))
				io.WriteString(logWriter, panicInfo)
				fn(c, logWriter.Bytes())

				if originWriter.GetStatusCode() == 0 {
					originWriter.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()

		c.Next()

		c.Writer = originWriter
		_writeResponseLog(c, logWriter, rec.Code, rec.Header(), rec.Body)

		end := time.Now()
		latency := end.Sub(start)

		szLatency := fmt.Sprintf("latency:%5.3fms\r\n", float64(latency)/float64(time.Millisecond))

		io.WriteString(logWriter, "-----------------------\r\n")
		io.WriteString(logWriter, szLatency)

		io.WriteString(logWriter, "----errors below-------\r\n")
		for _, err := range c.GetErrors() {
			if errpos, ok := err.(ErrorWithPos); ok {
				io.WriteString(logWriter, errpos.String()+"\r\n\r\n")
			} else {
				io.WriteString(logWriter, err.Error()+"\r\n")
			}
		}

		io.WriteString(logWriter, "-----infos below-------\r\n")
		for _, szInfo := range c.GetLogs() {
			io.WriteString(logWriter, szInfo+"\r\n")
		}

		io.WriteString(logWriter, "<----------------------LOG END----------------------------->\r\n")
		fn(c, logWriter.Bytes())
	}
}
