package fmx

import (
	"bytes"
	"errors"
	"net/http"
)

type IWriter interface {
	http.ResponseWriter
	GetStatusCode() int
	Init(w http.ResponseWriter)
	EnableRecordBody()
	GetRespBody() []byte
}

func NewWriter(w http.ResponseWriter) IWriter {
	return &writerImpl{
		realwriter:  w,
		statusCode:  0,
		bRecordBody: false,
		body:        nil,
	}
}

type writerImpl struct {
	realwriter  http.ResponseWriter
	statusCode  int
	bRecordBody bool
	body        *bytes.Buffer
}

func (this *writerImpl) Header() http.Header {
	return this.realwriter.Header()
}

func (this *writerImpl) Write(p []byte) (int, error) {
	if this.statusCode == 0 {
		panic(errors.New("status code not set"))
	}

	if this.bRecordBody {
		if this.body == nil {
			this.body = new(bytes.Buffer)
		}

		this.body.Write(p)
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

func (this *writerImpl) GetRespBody() []byte {
	if this.body == nil {
		return nil
	}

	return this.body.Bytes()
}

func (this *writerImpl) EnableRecordBody() {
	this.bRecordBody = true
}

func (this *writerImpl) Init(w http.ResponseWriter) {
	this.realwriter = w
	this.statusCode = 0
	this.bRecordBody = false

	if this.body != nil {
		this.body.Reset()
	}

	return
}
