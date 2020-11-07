package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
)

var (
	proxy_addr = "http://127.0.0.1:2003"
	port       = "2002"
)

func main() {
	http.HandleFunc("/", handler)
	log.Println("Start serving on port" + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
}

func handler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println(request.URL.RawQuery)
	fmt.Println(request.URL.Query())
	fmt.Println(request.URL.RawPath)
	fmt.Println(request.URL.Path)
	fmt.Println(request.URL.Host)

	url, err := url.Parse("http://127.0.0.1:2002/hello%2fzzz?name=laji")
	fmt.Println(url.RawQuery)
	fmt.Println(url.Query())
	fmt.Println(url.RawPath)
	fmt.Println(url.Path)
	fmt.Println(url.Host)
	proxy, err := url.Parse(proxy_addr)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	request.URL.Scheme = proxy.Scheme
	request.URL.Host = proxy.Host

	transport := http.DefaultTransport
	resp, err := transport.RoundTrip(request)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	resp.Write(writer)

}
