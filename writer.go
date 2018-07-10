package fmx

import (
	"errors"
	"net/http"
)

type IWriter interface {
	http.ResponseWriter
	GetStatusCode() int
	Init(w http.ResponseWriter)
}

func NewWriter(w http.ResponseWriter) IWriter {
	return &writerImpl{
		realwriter: w,
		statusCode: 0,
	}
}

type writerImpl struct {
	realwriter http.ResponseWriter
	statusCode int
}

func (this *writerImpl) Header() http.Header {
	return this.realwriter.Header()
}

func (this *writerImpl) Write(p []byte) (int, error) {
	if this.statusCode == 0 {
		panic(errors.New("status code not set"))
	}

	return this.realwriter.Write(p)
}

func (this *writerImpl) WriteHeader(status int) {
	if this.statusCode != 0 {
		panic(errors.New("HTTP Headers were already written!"))
	}

	this.statusCode = status
	this.realwriter.WriteHeader(status)
}

func (this *writerImpl) GetStatusCode() int {
	return this.statusCode
}

func (this *writerImpl) Init(w http.ResponseWriter) {
	this.realwriter = w
	this.statusCode = 0
	return
}
