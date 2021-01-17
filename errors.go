package fmx

import (
	"fmt"
	"runtime"
	"strings"
)

type ErrorWithPos interface {
	error
	Where() []string
	String() string
}

func Errorf(code int, format string, a ...interface{}) ErrorWithPos {
	e := &errorString{}
	e.s = fmt.Sprintf(format, a...)
	e.pos = make([]string, 0, 5)
	e.code = code

	pc, file, lineno, ok := runtime.Caller(1)
	if ok {
		e.pos = append(e.pos, fmt.Sprintf("%s:%d %s", file, lineno, runtime.FuncForPC(pc).Name()))
	}

	return e
}

func Error(err error, code ...int) ErrorWithPos {
	if err == nil {
		return nil
	}

	var stackinfo string
	pc, file, lineno, ok := runtime.Caller(1)
	if ok {
		stackinfo = fmt.Sprintf("%s:%d %s", file, lineno, runtime.FuncForPC(pc).Name())
	}

	if pErr, ok := err.(*errorString); ok {
		if len(stackinfo) > 0 {
			pErr.pos = append(pErr.pos, stackinfo)
		}

		if len(code) > 0 {
			pErr.code = code[0]
		}

		return pErr
	} else {
		e := &errorString{}
		e.s = err.Error()
		e.pos = make([]string, 0, 5)
		e.code = 0

		if len(stackinfo) > 0 {
			e.pos = append(e.pos, stackinfo)
		}

		if len(code) > 0 {
			e.code = code[0]
		}

		return e
	}
}

//append text description to err, return a new err
func ErrorAppend(err error, szExtra ...string) ErrorWithPos {
	if err == nil {
		return nil
	}

	var stackinfo string
	pc, file, lineno, ok := runtime.Caller(1)
	if ok {
		stackinfo = fmt.Sprintf("%s:%d %s", file, lineno, runtime.FuncForPC(pc).Name())
	}

	e := &errorString{}
	e.s = err.Error()
	e.pos = make([]string, 0, 5)
	e.code = 0

	if len(szExtra) > 0 {
		e.s += strings.Join(szExtra, "")
	}

	if pErr, ok := err.(*errorString); ok {
		e.pos = append(e.pos, pErr.pos...)
		e.code = pErr.code
	}

	if len(stackinfo) > 0 {
		e.pos = append(e.pos, stackinfo)
	}

	return e
}

func ErrCode(err error, def ...int) int {
	var code int = 0

	if len(def) > 0 {
		code = def[0]
	}

	if err == nil {
		return code
	}

	if pErr, ok := err.(*errorString); ok {
		return pErr.code
	}

	return code
}

type errorString struct {
	pos  []string
	s    string
	code int
}

func (e *errorString) Where() []string {
	return e.pos
}

func (e *errorString) Error() string {
	return e.s
}

func (e *errorString) String() string {

	szErr := e.s
	for _, item := range e.pos {
		szErr += "\r\n" + item
	}

	return szErr
}
