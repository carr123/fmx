package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/carr123/fmx"
)

func main() {
	router := fmx.New()
	router.Use() //fmx.SimpleLogger()

	router.GET("/api/profile", fmx.FullLogger(), GetProfile) //get json response
	router.GET("/api/export", ExportFile)                    //export file (web browser will download this file)
	router.POST("/api/profile", PostProfile)                 //client post json data to server
	router.POST("/api/avatar", PostImage)                    //post image to server through form data

	router.ServeDir("/", filepath.Join(getAppDir(), "www")) //server your static web pages

	fmt.Println("server start ...")
	fmt.Println("open your browser and navigate to: http://127.0.0.1:8080")
	err := http.ListenAndServe("127.0.0.1:8080", fmx.PanicHandler(router, func(s string) {
		fmt.Println("panic:", s)
	}))
	if err != nil {
		fmt.Println(err)
	}
}

func getAppDir() string {
	file, _ := exec.LookPath(os.Args[0])
	apppath, _ := filepath.Abs(file)
	dir := filepath.Dir(apppath)
	return dir
}
