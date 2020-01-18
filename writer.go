package fmx

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
)

type IWriter interface {
	http.ResponseWriter
	http.Hijacker
	GetStatusCode() int
	Init(w http.ResponseWriter)
	SetRecordRespBody(bool)
	GetRespBody() io.Reader
}

func NewWriter(w http.ResponseWriter) IWriter {
	return &WriterImpl{
		realwriter: w,
		statusCode: 0,
	}
}

type WriterImpl struct {
	realwriter      http.ResponseWriter
	statusCode      int
	bRecordRespBody bool
	respBody        bytes.Buffer
}

func (this *WriterImpl) Header() http.Header {
	return this.realwriter.Header()
}

func (this *WriterImpl) Write(p []byte) (int, error) {
	if this.statusCode == 0 {
		panic(errors.New("status code not set"))
	}

	var w io.Writer
	if this.bRecordRespBody {
		w = io.MultiWriter(&this.respBody, this.realwriter)
	} else {
		w = this.realwriter
	}

	return w.Write(p)
}

func (this *WriterImpl) WriteHeader(status int) {
	if this.statusCode != 0 {
		panic(errors.New("HTTP Headers were already written!"))
	}

	this.statusCode = status
	this.realwriter.WriteHeader(status)
}

func (this *WriterImpl) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := this.realwriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("doesn't support hijacking")
	}
	return hijacker.Hijack()
}

func (this *WriterImpl) GetStatusCode() int {
	return this.statusCode
}

func (this *WriterImpl) Init(w http.ResponseWriter) {
	this.realwriter = w
	this.statusCode = 0
	this.bRecordRespBody = false
	this.respBody.Reset()
	return
}

func (this *WriterImpl) SetRecordRespBody(bRecordBody bool) {
	this.bRecordRespBody = bRecordBody
}

func (this *WriterImpl) GetRespBody() io.Reader {
	return &this.respBody
}
