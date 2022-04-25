package fmx

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/CloudyKit/router"
)

const (
	Version string = "1.0.3"
)

type Engine struct {
	*Router
	httprouter *router.Router
	pool       sync.Pool
}

func NewServeMux() *Engine {
	engine := &Engine{}
	engine.Router = &Router{
		handlers: nil,
		engine:   engine,
	}

	engine.httprouter = router.New()
	engine.pool.New = func() interface{} {
		c := &Context{}
		c.Writer = &WriterImpl{}
		c.index = -1
		return c
	}

	return engine
}

func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	apppath, _ := filepath.Abs(file)
	dir := filepath.Dir(apppath)
	return dir
}
