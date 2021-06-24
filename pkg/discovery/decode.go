package discovery

import (
	"fmt"
	"strings"
)

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/remoting"
)

type (
	// DecodeServiceFunc 将原始数据解码为Service事件
	DecodeServiceFunc func(bytes []byte) (service flux.ServiceSpec, err error)

	// DecodeEndpointFunc 将原始数据解码为Service事件
	DecodeEndpointFunc func(bytes []byte) (endpoint flux.EndpointSpec, err error)

	// ServiceFilter 过滤和处理Service
	ServiceFilter func(event remoting.NodeEvent, data *flux.ServiceSpec) bool

	// EndpointFilter 过滤和重Endpoint
	EndpointFilter func(event remoting.NodeEvent, data *flux.EndpointSpec) bool
)

func VerifyJSON(bytes []byte) error {
	size := len(bytes)
	if size < len("{\"k\":0}") {
		return fmt.Errorf("check json: malformed, size: %d", size)
	}
	prefix := strings.TrimSpace(string(bytes[:5]))
	if prefix[0] != '[' && prefix[0] != '{' {
		return fmt.Errorf("check json: malformed, token: %s", prefix)
	}
	return nil
}
