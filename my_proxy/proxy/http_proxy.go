package proxy

import (
	"bytes"
	"fmt"
	"gateway/my_proxy/load_balance"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var transport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   time.Second * 30,
		KeepAlive: time.Second * 30,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       time.Second * 90,
	TLSHandshakeTimeout:   time.Second * 10,
	ExpectContinueTimeout: time.Second,
}

func NewHttpProxy(ld load_balance.LoadBalance) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		nextAddr, err := ld.Get("")
		if err != nil {
			log.Fatal(err)
		}
		target, err := url.Parse(nextAddr)
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path, req.URL.RawPath = joinURLPath(req.URL, target)
		req.Host = target.Host
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "& " + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "user-agent")
		}
	}

	modifyResponse := func(res *http.Response) error {
		var payload []byte

		if res.StatusCode != 200 {
			payload = []byte("StatueCode error :" + string(res.StatusCode))
		}

		res.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
		res.ContentLength = int64(len(payload))
		res.Header.Set("Content-Length", strconv.FormatInt(int64(len(payload)), 10))
		return nil
	}

	var errorHandler = func(res http.ResponseWriter, req *http.Request, e error) {
		fmt.Println(e)
	}

	return &httputil.ReverseProxy{
		Director:       director,
		Transport:      transport,
		ModifyResponse: modifyResponse,
		ErrorHandler:   errorHandler,
	}

}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}

	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
