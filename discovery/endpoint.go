package discovery

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/remoting"
	"strings"
)

var (
	emptyEndpointEvent = flux.EndpointEvent{}
)

type CompatibleEndpoint struct {
	flux.Endpoint
	RefService  flux.Service           `json:"service"`
	Authorize   bool                   `json:"authorize"`
	Extensions  map[string]interface{} `json:"extensions"`
	Permission  flux.Service           `json:"permission" yaml:"permission"`   // 权限服务定义
	Permissions []string               `json:"permissions" yaml:"permissions"` // 多组权限验证服务ID列表
}

func NewEndpointEvent(bytes []byte, etype remoting.EventType) (fxEvt flux.EndpointEvent, err error) {
	// Check json text
	size := len(bytes)
	if size < len("{\"k\":0}") {
		return emptyEndpointEvent, fmt.Errorf("ILLEGAL_JSONSIZE: %d", size)
	}
	prefix := strings.TrimSpace(string(bytes[:5]))
	if prefix[0] != '[' && prefix[0] != '{' {
		return emptyEndpointEvent, fmt.Errorf("ILLEGAL_JSONDATA: %s", string(bytes))
	}
	comp := CompatibleEndpoint{}
	if err := ext.JSONUnmarshal(bytes, &comp); nil != err {
		return emptyEndpointEvent, fmt.Errorf("ILLEGAL_JSONFORMAT: err: %w", err)
	}
	// 兼容旧结构: before V0.10
	if len(comp.Attributes) == 0 && comp.Kind == "" {
		// 1. Authorize
		comp.Attributes = []flux.Attribute{
			{Name: flux.EndpointAttrTagAuthorize, Value: comp.Authorize},
		}
		// 2. Extension to attributes
		for k, v := range comp.Extensions {
			comp.Attributes = append(comp.Attributes, flux.Attribute{Name: k, Value: v})
		}
	}

	// 兼容旧结构: before v0.18
	// Remove Endpoint.Service, use Endpoint.ServiceId instead
	if comp.Kind == "" {
		// 1. Service id
		if comp.ServiceId == "" {
			comp.ServiceId = comp.RefService.ServiceId
		}
		// 2. Permission service id
		if comp.Permission.IsValid() {
			comp.Attributes = append(comp.Attributes, flux.Attribute{
				Name: flux.FeaturePermission, Value: comp.Permission.ServiceID(),
			})
		}
		// 3. PermissionIds
		for _, id := range comp.Permissions {
			comp.Attributes = append(comp.Attributes, flux.Attribute{
				Name: flux.FeaturePermission, Value: id,
			})
		}
	}

	// 检查有效性
	if !comp.IsValid() {
		return emptyEndpointEvent, fmt.Errorf("INVALID_VALUES: data=%s", string(bytes))
	}

	event := flux.EndpointEvent{Endpoint: comp.Endpoint}
	switch etype {
	case remoting.EventTypeNodeAdd:
		event.EventType = flux.EventTypeAdded
	case remoting.EventTypeNodeDelete:
		event.EventType = flux.EventTypeRemoved
	case remoting.EventTypeNodeUpdate:
		event.EventType = flux.EventTypeUpdated
	default:
		return emptyEndpointEvent, fmt.Errorf("UNKNOWN_EVT_TYPE: type=%d", etype)
	}
	return event, nil
}
