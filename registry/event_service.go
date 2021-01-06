package registry

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/remoting"
)

var (
	invalidBackendServiceEvent = flux.BackendServiceEvent{}
)

type compatibleBackendService struct {
	flux.BackendService
	RpcProto   string `json:"rpcProto"`   // Service侧的协议
	RpcGroup   string `json:"rpcGroup"`   // Service侧的接口分组
	RpcVersion string `json:"rpcVersion"` // Service侧的接口版本
	RpcTimeout string `json:"rpcTimeout"` // Service侧的调用超时
	RpcRetries string `json:"rpcRetries"` // Service侧的调用重试
}

func toBackendServiceEvent(bytes []byte, etype remoting.EventType) (fxEvt flux.BackendServiceEvent, ok bool) {
	// Check json text
	size := len(bytes)
	if size < len("{\"k\":0}") || (bytes[0] != '[' && bytes[size-1] != '}') {
		logger.Infow("Invalid service event data.size", "data", string(bytes))
		return invalidBackendServiceEvent, false
	}
	comp := compatibleBackendService{}
	if err := ext.JSONUnmarshal(bytes, &comp); nil != err {
		logger.Warnw("Invalid service data",
			"event-type", etype, "data", string(bytes), "error", err)
		return invalidBackendServiceEvent, false
	}
	logger.Infow("Received service event",
		"event-type", etype, "service-id", comp.ServiceId, "data", string(bytes))
	// 检查有效性
	if !comp.IsValid() {
		logger.Warnw("illegal backend service", "data", string(bytes))
		return invalidBackendServiceEvent, false
	}
	// 兼容旧协议数据格式
	if len(comp.BackendService.Attributes) == 0 {
		comp.Attributes = []flux.Attribute{
			{
				Tag:   flux.ServiceAttrTagRpcProto,
				Name:  "RpcProto",
				Value: comp.RpcProto,
			},
			{
				Tag:   flux.ServiceAttrTagRpcGroup,
				Name:  "RpcGroup",
				Value: comp.RpcGroup,
			},
			{
				Tag:   flux.ServiceAttrTagRpcVersion,
				Name:  "RpcVersion",
				Value: comp.RpcVersion,
			},
			{
				Tag:   flux.ServiceAttrTagRpcRetries,
				Name:  "RpcRetries",
				Value: comp.RpcRetries,
			},
			{
				Tag:   flux.ServiceAttrTagRpcTimeout,
				Name:  "RpcTimeout",
				Value: comp.RpcTimeout,
			},
		}
	}
	event := flux.BackendServiceEvent{
		Service: comp.BackendService,
	}
	switch etype {
	case remoting.EventTypeNodeAdd:
		event.EventType = flux.EventTypeAdded
	case remoting.EventTypeNodeDelete:
		event.EventType = flux.EventTypeRemoved
	case remoting.EventTypeNodeUpdate:
		event.EventType = flux.EventTypeUpdated
	default:
		return invalidBackendServiceEvent, false
	}
	return event, true
}
