package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"github.com/dfcfw/spdy"
	"io"
	"log"
	"net"
	"net/http"
)

func main() {
	under, err := net.Dial("tcp", "127.0.0.1:6666")
	if err != nil {
		log.Fatal(err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer under.Close()

	// 发送 HTTP 握手
	req, _ := http.NewRequest(http.MethodConnect, "http://127.0.0.1:6666", nil)
	if err = req.Write(under); err != nil {
		log.Fatalf("HTTP 请求发送失败：%v", err)
	}
	res, err := http.ReadResponse(bufio.NewReader(under), req)
	if err != nil {
		log.Fatalf("HTTP 读取响应失败：%v", err)
	}
	if code := res.StatusCode; code != http.StatusOK {
		log.Fatalf("HTTP 协商失败：%d", code)
	}
	dat, _ := io.ReadAll(res.Body)
	log.Printf("服务端响应信息：%s", dat)

	mux := spdy.Client(under)
	dialFn := func(ctx context.Context, network, addr string) (net.Conn, error) {
		if network == "tcp" && addr == "em.com:80" {
			return mux.Dial()
		}
		return nil, &net.AddrError{Addr: addr}
	}

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: dialFn,
		},
	}

	req1, _ := http.NewRequest(http.MethodGet, "http://em.com/ping", nil)
	res1, err := cli.Do(req1)
	log.Println(res1)
	log.Println(err)

	body := map[string]string{"uname": "admin", "passwd": "123456"}
	raw, _ := json.Marshal(body)
	rc := io.NopCloser(bytes.NewReader(raw))
	req2, _ := http.NewRequest(http.MethodPost, "http://em.com/login", rc)
	res2, err := cli.Do(req2)
	log.Println(res2)
	log.Println(err)

}
