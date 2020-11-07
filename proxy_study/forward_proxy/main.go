package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

type Pxy struct {
}

func (p Pxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	fmt.Printf("Receuved request %s %s %s \n", request.Method, request.Host, request.RemoteAddr)
	fmt.Println(request.Header)
	transport := http.DefaultTransport

	outReq := new(http.Request)
	*outReq = *request
	if clientIP, _, err := net.SplitHostPort(request.RemoteAddr); err == nil {
		if prior, ok := outReq.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ",") + "," + clientIP
		}
		outReq.Header.Set("X-Forwardef-For", clientIP)
	}
	fmt.Println(outReq.Header)

	res, err := transport.RoundTrip(outReq)
	if err != nil {
		writer.WriteHeader(http.StatusBadGateway)
		return
	}

	for key, value := range res.Header {
		for _, v := range value {
			writer.Header().Add(key, v)
		}
	}
	fmt.Println("url", request.URL)
	writer.WriteHeader(res.StatusCode)
	io.Copy(writer, res.Body)
	res.Body.Close()

}

func main() {
	pxy := Pxy{}
	http.Handle("/", pxy)
	http.ListenAndServe("0.0.0.0:1234", nil)

	net.Listen("tcp", ":1234")
}
