package discovery

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/logger"
	"github.com/bytepowered/flux/flux-node/remoting"
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

func NewBackendServiceEvent(bytes []byte, etype remoting.EventType, node string) (fxEvt flux.BackendServiceEvent, ok bool) {
	// Check json text
	size := len(bytes)
	if size < len("{\"k\":0}") || (bytes[0] != '[' && bytes[size-1] != '}') {
		logger.Warnw("DISCOVERY:SERVICE:ILLEGAL_JSONSIZE", "data", string(bytes), "node", node)
		return invalidBackendServiceEvent, false
	}
	service := flux.BackendService{}
	if err := ext.JSONUnmarshal(bytes, &service); nil != err {
		logger.Warnw("DISCOVERY:SERVICE:ILLEGAL_JSONFORMAT",
			"event-type", etype, "data", string(bytes), "error", err, "node", node)
		return invalidBackendServiceEvent, false
	}
	// 检查有效性
	if !service.IsValid() {
		logger.Warnw("DISCOVERY:SERVICE:INVALID_VALUES", "service", service, "node", node)
		return invalidBackendServiceEvent, false
	}
	EnsureServiceAttrs(&service)

	event := flux.BackendServiceEvent{Service: service}
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

// EnsureServiceAttrs 兼容旧协议数据格式
func EnsureServiceAttrs(service *flux.BackendService) *flux.BackendService {
	if len(service.Attributes) == 0 {
		service.Attributes = []flux.Attribute{
			{Name: flux.ServiceAttrTagRpcProto, Value: service.RpcProto},
			{Name: flux.ServiceAttrTagRpcGroup, Value: service.RpcGroup},
			{Name: flux.ServiceAttrTagRpcVersion, Value: service.RpcVersion},
			{Name: flux.ServiceAttrTagRpcRetries, Value: service.RpcRetries},
			{Name: flux.ServiceAttrTagRpcTimeout, Value: service.RpcTimeout},
		}
	}
	return service
}
