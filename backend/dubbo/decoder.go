package dubbo

import (
	"github.com/apache/dubbo-go/protocol"
	"github.com/bytepowered/flux"
	"net/http"
)

const (
	ResponseKeyStatusCode = "@net.bytepowered.flux.http-status"
	ResponseKeyHeaders    = "@net.bytepowered.flux.http-headers"
	ResponseKeyBody       = "@net.bytepowered.flux.http-body"
)

func NewBackendResultDecodeFuncWith(codeKey, headerKey, bodyKey string) flux.BackendResultDecodeFunc {
	return func(ctx flux.Context, dubbores interface{}) (*flux.BackendResult, error) {
		data := dubbores
		attachments := make(map[string]interface{}, 8)
		// 支持Dubbo返回Result类型
		if rpcr, ok := dubbores.(protocol.Result); ok {
			if err := rpcr.Error(); nil != err {
				return nil, err
			}
			data = rpcr.Result()
			for k, v := range rpcr.Attachments() {
				attachments[k] = v
			}
		}
		values, ismapv := WrapBodyValues(data)
		if !ismapv {
			return &flux.BackendResult{
				StatusCode:  flux.StatusOK,
				Headers:     make(http.Header, 0),
				Attachments: attachments,
				Body:        dubbores,
			}, nil
		}
		// Header
		header, err := values.ReadHeaderValue(headerKey)
		if nil != err {
			return nil, err
		}
		// StatusCode
		status, err := values.ReadStatusValue(codeKey)
		if nil != err {
			return nil, err
		}
		// Body
		return &flux.BackendResult{
			StatusCode:  status,
			Headers:     header,
			Body:        values.ReadBodyValue(bodyKey),
			Attachments: attachments,
		}, nil
	}
}

func NewBackendResultDecodeFunc() flux.BackendResultDecodeFunc {
	return NewBackendResultDecodeFuncWith(ResponseKeyStatusCode, ResponseKeyHeaders, ResponseKeyBody)
}
