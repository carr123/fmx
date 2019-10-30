package fmx

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"runtime"
	"strconv"

	"github.com/CloudyKit/router"
)

const (
	contentType    = "Content-Type"
	acceptLanguage = "Accept-Language"
	AbortIndex     = math.MaxInt16 / 2
)

var jsonContentType = []string{"application/json; charset=utf-8"}
var plainContentType = []string{"text/plain; charset=utf-8"}

type HandlerFunc func(*Context)
type H map[string]interface{}

type Context struct {
	Writer   IWriter
	Request  *http.Request
	Keys     map[string]interface{}
	HasError bool
	params   *router.Parameter
	handlers []HandlerFunc
	index    int16
	_errs    []error
	_logs    []string
}

func (engine *Engine) createContext(w http.ResponseWriter, r *http.Request) *Context {
	c := engine.pool.Get().(*Context)
	c.Writer.Init(w)
	c.Request = r
	c.index = -1
	c.Keys = nil
	c._errs = c._errs[:0]
	c._logs = c._logs[:0]
	c.HasError = false

	return c
}

func (c *Context) ReadBody() []byte {
	bin, _ := ioutil.ReadAll(c.Request.Body)
	return bin
}

func (c *Context) Next() {
	c.index++
	s := int16(len(c.handlers))
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Abort() {
	c.index = AbortIndex
}

func (c *Context) AbortWithStatus(code int) {
	c.index = AbortIndex
	c.Writer.WriteHeader(code)
}

// Set is used to store a new key/value pair exclusivelly for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *Context) Set(key string, value interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (c *Context) Get(key string) (value interface{}, exists bool) {
	if c.Keys != nil {
		value, exists = c.Keys[key]
	}
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (c *Context) MustGet(key string) interface{} {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

//Param get param from route
func (c *Context) Param(name string) string {
	return c.params.ByName(name)
}

func (c *Context) ClientIP() string {
	requester := c.Request.Header.Get("X-Real-IP")

	if len(requester) == 0 {
		requester = c.Request.Header.Get("X-Forwarded-For")
	}

	if len(requester) == 0 {
		requester = c.Request.RemoteAddr
	}

	return requester
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}

func (c *Context) JSON(code int, v interface{}) {
	writeContentType(c.Writer, jsonContentType)
	c.Writer.WriteHeader(code)

	if v != nil {
		if err := json.NewEncoder(c.Writer).Encode(v); err != nil {
			panic(err)
		}
	}
}

func (c *Context) String(code int, format string, values ...interface{}) {
	writeContentType(c.Writer, plainContentType)
	c.Writer.WriteHeader(code)

	if len(values) > 0 {
		fmt.Fprintf(c.Writer, format, values...)
	} else {
		io.WriteString(c.Writer, format)
	}
}

// Data writes some data into the body stream and updates the HTTP code.
func (c *Context) Data(code int, contentType string, data []byte) {
	writeContentType(c.Writer, []string{contentType})
	c.Writer.WriteHeader(code)

	c.Writer.Write(data)
}

func (c *Context) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}

func (c *Context) Redirect(code int, Location string) {
	http.Redirect(c.Writer, c.Request, Location, code)
}

func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) QueryString(key, def string) string {
	if values, ok := c.Request.URL.Query()[key]; ok && len(values) > 0 {
		return values[0]
	}

	return def
}

func (c *Context) QueryInt(key string, def int) int {
	val := c.Request.URL.Query().Get(key)
	if val == "" {
		return def
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}

	return i
}

func (c *Context) QueryUint(key string, def uint) uint {
	val := c.Request.URL.Query().Get(key)
	if val == "" {
		return def
	}

	i, err := strconv.ParseUint(val, 10, 0)
	if err != nil {
		panic(err)
	}

	return uint(i)
}

func (c *Context) QueryInt64(key string, def int64) int64 {
	val := c.Request.URL.Query().Get(key)
	if val == "" {
		return def
	}

	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		panic(err)
	}

	return i
}

func (c *Context) QueryFloat64(key string, def float64) float64 {
	val := c.Request.URL.Query().Get(key)
	if val == "" {
		return def
	}

	f, err := strconv.ParseFloat(c.Request.URL.Query().Get(key), 64)
	if err != nil {
		panic(err)
	}

	return f
}

func (c *Context) QueryByte(key string, def byte) byte {
	val := c.Request.URL.Query().Get(key)
	if val == "" {
		return def
	}

	i, err := strconv.ParseUint(val, 10, 8)
	if err != nil {
		panic(err)
	}

	return byte(i)
}

/*
func (c *Context) AddError(err ...error) {
	for _, item := range err {
		if item != nil {
			c._errs = append(c._errs, item)
			c.HasError = true
		}
	}
}
*/

func (c *Context) AddError(err error) {
	if err == nil {
		return
	}

	var stackinfo string
	pc, file, lineno, ok := runtime.Caller(1)
	if ok {
		stackinfo = fmt.Sprintf("%s:%d %s", file, lineno, runtime.FuncForPC(pc).Name())
	}

	if _, ok := err.(*errorString); !ok {
		e := &errorString{}
		e.s = err.Error()
		e.pos = make([]string, 0, 5)
		e.code = 0

		if len(stackinfo) > 0 {
			e.pos = append(e.pos, stackinfo)
		}
		c._errs = append(c._errs, e)
	} else {
		c._errs = append(c._errs, err)
	}
}

func (c *Context) AddLog(info ...string) {
	c._logs = append(c._logs, info...)
}

func (c *Context) GetErrors() []error {
	return c._errs
}

func (c *Context) GetLogs() []string {
	return c._logs
}
