package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/astaxie/beego/logs"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			logs.Error(err)
			return
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/test", NewTest())
	logs.Info("http服务器启动，端口：8084")
	err := http.ListenAndServe(":8084", mux)
	if err != nil {
		logs.Error("启动失败", err)
	}
}

type test struct{}

func NewTest() *test {
	return &test{}
}

func (t *test) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")                           //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type,Authorization") //header的类型
	logs.Info(`有人访问`)
	if r.Method != "POST" {
		err := fmt.Errorf(`the method is should be POST`)
		logs.Error(err)
		return
	}
	file, header, err := r.FormFile(`file`)
	if err != nil {
		logs.Error(err)
		return
	}
	logs.Info(header.Filename)
	newFile, err := os.Create(`./` + header.Filename)
	if err != nil {
		logs.Error(err)
		return
	}
	defer newFile.Close()
	_, err = io.Copy(newFile, file)
	if err != nil {
		logs.Error(err)
		return
	}

	w.Write([]byte("upload success"))
}
