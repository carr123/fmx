package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/carr123/fmx"
	"golang.org/x/net/websocket"
)

func main() {
	router := fmx.New()
	router.Use(fmx.Logger(true, true)) //

	//get json response
	router.GET("/api/profile", func(c *fmx.Context) {
		name := c.Query("name")
		fmt.Println("name:", name)

		c.AddLog(fmt.Sprintf("query user %s", name))

		//c.Writer.WriteHeader(200)
		//c.Writer.Write([]byte("haha"))
		//return

		if name == "jack" {
			c.JSON(200, fmx.H{"name": name, "age": 20})
			return
		} else {
			c.AddError(fmx.NewError("username error", 400))
			c.String(400, "username error")
			return
		}
	})

	//export file (web browser will download this file)
	router.GET("/api/export", func(c *fmx.Context) {
		name := c.Query("name")
		if name != "jack" {
			c.String(400, "username error")
			return
		}

		var output io.Writer
		if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Writer.Header().Set("Content-Encoding", "gzip")
			zipw := gzip.NewWriter(c.Writer)
			defer zipw.Close()
			output = zipw
		} else {
			output = c.Writer
		}

		c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s.txt", name))
		c.Writer.Header().Set("Content-Type", "application/octet-stream")

		filecontent := fmt.Sprintf("name=%s age=%d", name, 20)

		c.Writer.WriteHeader(200)
		io.Copy(output, strings.NewReader(filecontent))
	})

	//client post json data to server
	router.POST("/api/profile", func(c *fmx.Context) {
		var User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		data := c.ReadBody()
		json.Unmarshal(data, &User)

		fmt.Println("you post:", User.Name, " ", User.Age)

		c.JSON(200, fmx.H{"success": true})
	})

	//post image to server through form data
	router.POST("/api/avatar", func(c *fmx.Context) {
		r := c.Request
		r.ParseMultipartForm(32 << 20)

		username := r.Form.Get("name")
		fmt.Println("name:", username)

		fimg, handler, err := r.FormFile("avatar")
		if err != nil {
			c.String(400, err.Error())
			return
		}

		defer fimg.Close()

		//save avatar to file
		fullPath := filepath.Join("e:/avatars/", username+filepath.Ext(handler.Filename))
		os.MkdirAll(filepath.Dir(fullPath), 0777)

		f, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			c.String(500, err.Error())
			return
		}

		defer f.Close()

		io.Copy(f, fimg)

		c.JSON(200, fmx.H{"success": true, "msg": "upload image success"})
	})

	notFound := func(c *fmx.Context) {
		c.String(404, "your web page is gone !!!")
	}

	pagerouter := router.Group(fmx.Code404Handler(notFound))
	pagerouter.ServeDir("/static", filepath.Join(getAppDir(), "www"))

	router.GET("/ws/long", HandleWS())

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

func HandleWS() func(c *fmx.Context) {
	return func(c *fmx.Context) {
		s := websocket.Server{websocket.Config{}, nil, func(ws *websocket.Conn) {
			defer ws.Close()
			var msgRead = make([]byte, 1024*1024)
			var n int
			var err error

			fmt.Printf("ws new client:%s\r\n", c.ClientIP())

			for {
				if n, err = ws.Read(msgRead); err != nil || n < 0 {
					break
				}
				fmt.Printf("websocket read from %s, data:%s\r\n", c.ClientIP(), string(msgRead[:n]))

				if err := websocket.Message.Send(ws, string(msgRead[:n])); err != nil {
					break
				}
			}
		}}

		s.ServeHTTP(c.Writer, c.Request)
	}
}
