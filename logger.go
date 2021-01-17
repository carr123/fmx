package fmx

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type LoggerFunc func(c *Context, szLog []byte)

func Logger(bShowReqBody bool, bShowRespBody bool) HandlerFunc {
	return LoggerWithFunc(bShowReqBody, bShowRespBody, DefaultLoggerFunc())
}

func LoggerWithFunc(bShowReqBody bool, bShowRespBody bool, fn LoggerFunc) HandlerFunc {
	if fn == nil {
		return func(c *Context) {
		}
	}

	_writeRequestLog := func(c *Context, t time.Time, logwriter io.Writer) {
		szIP := c.ClientIP()
		szProto := c.Request.Proto
		rawPath := c.Request.URL.String()
		method := c.Request.Method

		var reqHeader bytes.Buffer
		c.Request.Header.Write(&reqHeader)

		io.WriteString(logwriter, "\r\n<----------------------LOG BEGIN----------------------------->\r\n")
		io.WriteString(logwriter, "requesttime:"+t.Format("2006-01-02 15:04:05")+"\r\n")
		io.WriteString(logwriter, "clientaddr:"+szIP+"\r\n\r\n")
		io.WriteString(logwriter, method+" "+rawPath+"  "+szProto+"\r\n")
		io.WriteString(logwriter, "Host:"+c.Request.Host+"\r\n")
		io.WriteString(logwriter, reqHeader.String()+"\r\n")

		if bShowReqBody {
			if c.Request.Body != nil {
				buff := &bytes.Buffer{}
				io.Copy(buff, c.Request.Body)
				c.Request.Body.Close()
				c.Request.Body = ioutil.NopCloser(buff)
				io.WriteString(logwriter, buff.String()+"\r\n")
			}
		} else {
			io.WriteString(logwriter, "<body not printed>\r\n")
		}
	}

	_writeResponseLog := func(c *Context, logwriter io.Writer) {
		io.WriteString(logwriter, "----------------\r\n")
		io.WriteString(logwriter, "statuscode:"+fmt.Sprintf("%03d", c.Writer.GetStatusCode())+"\r\n")
		c.Writer.Header().Write(logwriter)
		io.WriteString(logwriter, "\r\n")

		if bShowRespBody {
			io.Copy(logwriter, c.Writer.GetRespBody())
		} else {
			io.WriteString(logwriter, "<body not printed>")
		}
	}

	return func(c *Context) {
		start := time.Now()

		logWriter := &bytes.Buffer{}
		_writeRequestLog(c, start, logWriter)
		c.Writer.SetRecordRespBody(bShowRespBody)

		defer func() {
			if e := recover(); e != nil {
				io.WriteString(logWriter, "--------------------\r\n")
				stack := stack(3)
				panicInfo := fmt.Sprintf("PANIC: %s\r\n%s", e, stack)
				c.AddError(Errorf(0, panicInfo))
				io.WriteString(logWriter, panicInfo)
				fn(c, logWriter.Bytes())

				if c.Writer.GetStatusCode() == 0 {
					c.Writer.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()

		c.Next()

		_writeResponseLog(c, logWriter)

		end := time.Now()
		latency := end.Sub(start)
		szLatency := fmt.Sprintf("latency:%5.3fms\r\n", float64(latency)/float64(time.Millisecond))

		io.WriteString(logWriter, "\r\n-----------------------\r\n")
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

func DefaultLoggerFunc() LoggerFunc {
	return LoggerFunc(func(c *Context, szLog []byte) {
		io.Copy(os.Stdout, bytes.NewReader(szLog))
	})
}
