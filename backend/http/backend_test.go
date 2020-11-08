package http

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestBackend_HttpbinMethods(t *testing.T) {
	backend := NewHttpBackend()
	for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
		inReq, _ := http.NewRequest(method, "http://localhost:8080/"+method, nil)
		inReq.GetBody = func() (io.ReadCloser, error) {
			if "DELETE" == method {
				return ioutil.NopCloser(strings.NewReader("callMethod=DELETE&paramFrom=BODY")), nil
			} else {
				return ioutil.NopCloser(bytes.NewReader([]byte{})), nil
			}
		}
		bodyReader, _ := inReq.GetBody()
		service := newHttpBinService(method)
		newReq, err := backend.Assemble(&service, inReq.URL, bodyReader, context.Background())
		if nil != err {
			t.Fatalf("invoke err: %s", err)
		}
		if err := testHttpRequest(newReq, checkerStatusCode200(t)); nil != err {
			t.Fatalf("request err: %s", err)
		}
	}

}

func newHttpBinService(method string) flux.BackendService {
	params := make([]flux.Argument, 0)
	if "DELETE" != method {
		params = []flux.Argument{
			{Name: "username", Value: flux.NewWrapValue("yongjia")},
			{Name: "email", Value: flux.NewWrapValue("yongjia.chen@hotmail.com")},
			{Name: "callMethod", Value: flux.NewWrapValue(method)},
			{Name: "paramFrom", Value: flux.NewWrapValue("endpoint-define")},
		}
	}
	return flux.BackendService{
		RpcTimeout: "10s",
		RemoteHost: "httpbin.org",
		Interface:  "/" + strings.ToLower(method),
		Method:     method,
		Arguments:  params,
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
