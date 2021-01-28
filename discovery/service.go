package discovery

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/remoting"
)

var (
	invalidBackendServiceEvent = flux.BackendServiceEvent{}
)

//
//type BackendService struct {
//	flux.BackendService
//	RpcProto   string `json:"rpcProto"`   // Service侧的协议
//	RpcGroup   string `json:"rpcGroup"`   // Service侧的接口分组
//	RpcVersion string `json:"rpcVersion"` // Service侧的接口版本
//	RpcTimeout string `json:"rpcTimeout"` // Service侧的调用超时
//	RpcRetries string `json:"rpcRetries"` // Service侧的调用重试
//}

func NewBackendServiceEvent(bytes []byte, etype remoting.EventType) (fxEvt flux.BackendServiceEvent, ok bool) {
	// Check json text
	size := len(bytes)
	if size < len("{\"k\":0}") || (bytes[0] != '[' && bytes[size-1] != '}') {
		logger.Warnw("Invalid service event data.size", "data", string(bytes))
		return invalidBackendServiceEvent, false
	}
	service := flux.BackendService{}
	if err := ext.JSONUnmarshal(bytes, &service); nil != err {
		logger.Warnw("Invalid service data",
			"event-type", etype, "data", string(bytes), "error", err)
		return invalidBackendServiceEvent, false
	}
	// 检查有效性
	if !service.IsValid() {
		logger.Warnw("illegal backend service", "service", service)
		return invalidBackendServiceEvent, false
	}
	setupServiceAttributes(&service)
	ensureServiceAttributeTagName(&service)
	event := flux.BackendServiceEvent{
		Service: service,
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

// setupServiceAttributes 兼容旧协议数据格式
func setupServiceAttributes(service *flux.BackendService) {
	if len(service.Attributes) == 0 {
		service.Attributes = []flux.Attribute{
			{
				Tag:   flux.ServiceAttrTagRpcProto,
				Name:  flux.ServiceAttrTagNames[flux.ServiceAttrTagRpcProto],
				Value: service.RpcProto,
			},
			{
				Tag:   flux.ServiceAttrTagRpcGroup,
				Name:  flux.ServiceAttrTagNames[flux.ServiceAttrTagRpcGroup],
				Value: service.RpcGroup,
			},
			{
				Tag:   flux.ServiceAttrTagRpcVersion,
				Name:  flux.ServiceAttrTagNames[flux.ServiceAttrTagRpcVersion],
				Value: service.RpcVersion,
			},
			{
				Tag:   flux.ServiceAttrTagRpcRetries,
				Name:  flux.ServiceAttrTagNames[flux.ServiceAttrTagRpcRetries],
				Value: service.RpcRetries,
			},
			{
				Tag:   flux.ServiceAttrTagRpcTimeout,
				Name:  flux.ServiceAttrTagNames[flux.ServiceAttrTagRpcTimeout],
				Value: service.RpcTimeout,
			},
		}
	}
}

func ensureServiceAttributeTagName(service *flux.BackendService) {
	// 订正Tag与Name的关系
	for i := range service.Attributes {
		ptr := &service.Attributes[i]
		newT, newName := flux.EnsureServiceAttribute(ptr.Tag, ptr.Name)
		ptr.Tag = newT
		ptr.Name = newName
	}
}
