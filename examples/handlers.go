package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/carr123/fmx"
)

func GetProfile(c *fmx.Context) {
	name := c.Query("name")
	fmt.Println("you query name:", name)

	c.AddLog(fmt.Sprintf("context log: query user %s", name))

	if name == "jack" {
		c.JSON(200, fmx.H{"name": name, "age": 20})
		return
	} else {
		c.AddError(fmx.NewError("username error", 400))
		c.String(400, "username error")
		return
	}
}

func ExportFile(c *fmx.Context) {
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
	c.Writer.WriteHeader(200)

	filecontent := fmt.Sprintf("name=%s age=%d", name, 20)
	io.Copy(output, strings.NewReader(filecontent))
}

func PostProfile(c *fmx.Context) {
	var User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := c.ReadBody()
	json.Unmarshal(data, &User)

	fmt.Println("you post:", string(data))

	c.JSON(200, fmx.H{"success": true})
}

func PostImage(c *fmx.Context) {
	r := c.Request
	r.ParseMultipartForm(32 << 20)

	username := r.Form.Get("name")

	fimg, handler, err := r.FormFile("avatar")
	if err != nil {
		c.String(400, err.Error())
		return
	}

	defer fimg.Close()

	//save avatar to file
	fullPath := filepath.Join(getAppDir(), "uploads", username+filepath.Ext(handler.Filename))
	os.MkdirAll(filepath.Dir(fullPath), 0777)

	f, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	defer f.Close()

	io.Copy(f, fimg)

	c.JSON(200, fmx.H{"success": true, "msg": fmt.Sprintf("your file saved to %s", fullPath)})
}
