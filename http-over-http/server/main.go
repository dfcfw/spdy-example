package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/dfcfw/spdy"
)

func main() {
	vit := new(virtual)
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", vit.Ping)
	mux.HandleFunc("/login", vit.Login)

	nrm := normal{virtual: mux}

	_ = http.ListenAndServe(":6666", nrm)
}

type virtual struct{}

func (virtual) Ping(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("PONG"))
	log.Println("-------[ PONG ]-------")
}

func (virtual) Login(w http.ResponseWriter, r *http.Request) {
	var user struct {
		Uname  string `json:"uname"`
		Passwd string `json:"passwd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("JSON Decoder 错误：%v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if user.Uname == "admin" && user.Passwd == "123456" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

type normal struct {
	virtual http.Handler
}

func (nm normal) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodConnect {
		log.Println("约定 CONNECT 方法才可以建立连接，请客户端使用 CONNECT 方法调用")
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Println("未实现 http.Hijacker 接口，请客户端使用 HTTP/1.1 协议")
		return
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("Hijack 发生错误：%v", err)
		return
	}

	// 返回一个 JSON Body
	dat := map[string]string{"msg": "升级成功"}
	raw, _ := json.Marshal(dat)

	code := http.StatusOK
	res := &http.Response{
		Status:        http.StatusText(code),
		StatusCode:    code,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
		Body:          io.NopCloser(bytes.NewReader(raw)),
		ContentLength: int64(len(raw)),
		Request:       r,
	}
	if err = res.Write(conn); err != nil {
		log.Printf("HTTP 响应写入失败：%v", err)
		return
	}

	// 通道建立成功，执行业务逻辑
	mux := spdy.Server(conn)
	srv := &http.Server{Handler: nm.virtual}
	if err = srv.Serve(mux); err != nil {
		log.Printf("Listen error: %v", err)
	}
}
