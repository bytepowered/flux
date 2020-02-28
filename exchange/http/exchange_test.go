package http

import (
	"bytes"
	"fmt"
	"github.com/bytepowered/flux"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestExchange_HttpbinMethods(t *testing.T) {
	ex := NewHttpExchange()
	for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
		inReq, _ := http.NewRequest(method, "http://localhost:8080/"+method, nil)
		inReq.GetBody = func() (io.ReadCloser, error) {
			if "DELETE" == method {
				return ioutil.NopCloser(strings.NewReader("callMethod=DELETE&paramFrom=BODY")), nil
			} else {
				return ioutil.NopCloser(bytes.NewReader([]byte{})), nil
			}
		}
		newReq, err := ex.Assemble(newHttpBinEndpoint(method), inReq)
		if nil != err {
			t.Fatalf("invoke err: %s", err)
		}
		if err := testHttpRequest(newReq, checkerStatusCode200(t)); nil != err {
			t.Fatalf("request err: %s", err)
		}
	}

}

func newHttpBinEndpoint(method string) *flux.Endpoint {
	params := make([]flux.Argument, 0)
	if "DELETE" != method {
		params = []flux.Argument{
			{ArgName: "username", ArgValue: flux.NewWrapValue("yongjia")},
			{ArgName: "email", ArgValue: flux.NewWrapValue("yongjia.chen@hotmail.com")},
			{ArgName: "callMethod", ArgValue: flux.NewWrapValue(method)},
			{ArgName: "paramFrom", ArgValue: flux.NewWrapValue("endpoint-define")},
		}
	}
	return &flux.Endpoint{
		RpcTimeout:     "10s",
		UpstreamHost:   "httpbin.org",
		UpstreamUri:    "/" + strings.ToLower(method),
		UpstreamMethod: method,
		Arguments:      params,
	}
}

func checkerStatusCode200(t *testing.T) func(resp *http.Response) {
	return func(resp *http.Response) {
		if 200 != resp.StatusCode {
			t.Fatalf("Status Code not OK: %d, url: %s", resp.StatusCode, resp.Request.URL)
		}
	}
}

func testHttpRequest(newReq *http.Request, checker ...func(resp *http.Response)) error {
	getResp, err := http.DefaultClient.Do(newReq)
	if nil != err {
		return err
	}
	for _, c := range checker {
		c(getResp)
	}
	fmt.Println(getResp.Status)
	fmt.Println()
	fmt.Println(getResp.Header)
	d, err := ioutil.ReadAll(getResp.Body)
	if nil != err {
		return err
	}
	fmt.Println()
	fmt.Println(string(d))
	return nil
}
