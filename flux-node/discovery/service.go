package discovery

import (
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/logger"
	"github.com/bytepowered/flux/flux-node/remoting"
)

var (
	invalidBackendServiceEvent = flux2.BackendServiceEvent{}
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

func NewBackendServiceEvent(bytes []byte, etype remoting.EventType, node string) (fxEvt flux2.BackendServiceEvent, ok bool) {
	// Check json text
	size := len(bytes)
	if size < len("{\"k\":0}") || (bytes[0] != '[' && bytes[size-1] != '}') {
		logger.Warnw("DISCOVERY:SERVICE:ILLEGAL_JSONSIZE", "data", string(bytes), "node", node)
		return invalidBackendServiceEvent, false
	}
	service := flux2.BackendService{}
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

	event := flux2.BackendServiceEvent{Service: service}
	switch etype {
	case remoting.EventTypeNodeAdd:
		event.EventType = flux2.EventTypeAdded
	case remoting.EventTypeNodeDelete:
		event.EventType = flux2.EventTypeRemoved
	case remoting.EventTypeNodeUpdate:
		event.EventType = flux2.EventTypeUpdated
	default:
		return invalidBackendServiceEvent, false
	}

	return event, true
}

// EnsureServiceAttrs 兼容旧协议数据格式
func EnsureServiceAttrs(service *flux2.BackendService) *flux2.BackendService {
	if len(service.Attributes) == 0 {
		service.Attributes = []flux2.Attribute{
			{Name: flux2.ServiceAttrTagRpcProto, Value: service.RpcProto},
			{Name: flux2.ServiceAttrTagRpcGroup, Value: service.RpcGroup},
			{Name: flux2.ServiceAttrTagRpcVersion, Value: service.RpcVersion},
			{Name: flux2.ServiceAttrTagRpcRetries, Value: service.RpcRetries},
			{Name: flux2.ServiceAttrTagRpcTimeout, Value: service.RpcTimeout},
		}
	}
	return service
}
