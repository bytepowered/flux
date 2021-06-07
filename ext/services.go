package ext

import (
	"github.com/bytepowered/flux"
	"sync"
)

var (
	serviceNotFound flux.Service
	services        = new(sync.Map)
)

func RegisterServiceByID(id string, service flux.Service) {
	flux.AssertTrue(len(id) > len("a:b"), "<service-id> malformed")
	services.Store(id, service)
}

// RegisterService store transporter service
func RegisterService(service flux.Service) {
	RegisterServiceByID(service.ServiceID(), service)
}

// Services 返回全部注册的Service
func Services() map[string]flux.Service {
	out := make(map[string]flux.Service, 128)
	services.Range(func(key, value interface{}) bool {
		out[key.(string)] = value.(flux.Service)
		return true
	})
	return out
}

// ServiceByID load transporter service by serviceId
func ServiceByID(serviceID string) (flux.Service, bool) {
	v, ok := services.Load(serviceID)
	if ok {
		return v.(flux.Service), true
	}
	return serviceNotFound, false
}

// RemoveServiceByID remove transporter service by serviceId
func RemoveServiceByID(serviceID string) {
	services.Delete(serviceID)
}

// HasServiceByID check service exists by service id
func HasServiceByID(serviceID string) bool {
	_, ok := services.Load(serviceID)
	return ok
}
