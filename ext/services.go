package ext

import (
	"fmt"
	"sync"

	"github.com/bytepowered/flux"
)

var (
	_serviceNotFound flux.BackendService
	_servicesMap     *sync.Map = new(sync.Map)
)

// StoreBackendService store backend service
func StoreBackendService(service flux.BackendService) {
	id := _ensureServiceID(&service)
	_servicesMap.Store(id, service)
}

// LoadBackendService load backend service by serviceId
func LoadBackendService(serviceID string) (flux.BackendService, bool) {
	v, ok := _servicesMap.Load(serviceID)
	if ok {
		return v.(flux.BackendService), true
	}
	return _serviceNotFound, false
}

// RemoveBackendService remove backend service by serviceId
func RemoveBackendService(serviceID string) {
	_servicesMap.Delete(serviceID)
}

// HasBackendService check service exists by service id
func HasBackendService(serviceID string) bool {
	_, ok := _servicesMap.Load(serviceID)
	return ok
}

func _ensureServiceID(service *flux.BackendService) string {
	id := service.ServiceId
	if "" == id {
		id = service.Interface + ":" + service.Method
	}
	if len(id) < len("a:b") {
		panic(fmt.Sprintf("BackendService must has an Id, service: %+v", service))
	}
	return id
}
