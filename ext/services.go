package ext

import (
	"fmt"
	"sync"

	"github.com/bytepowered/flux"
)

var (
	serviceNotFound flux.BackendService
	servicesMap     *sync.Map = new(sync.Map)
)

func StoreBackendServiceById(id string, service flux.BackendService) {
	servicesMap.Store(id, service)
}

// StoreBackendService store backend service
func StoreBackendService(service flux.BackendService) {
	id := _ensureServiceID(&service)
	StoreBackendServiceById(id, service)
}

// LoadBackendService load backend service by serviceId
func LoadBackendService(serviceID string) (flux.BackendService, bool) {
	v, ok := servicesMap.Load(serviceID)
	if ok {
		return v.(flux.BackendService), true
	}
	return serviceNotFound, false
}

// RemoveBackendService remove backend service by serviceId
func RemoveBackendService(serviceID string) {
	servicesMap.Delete(serviceID)
}

// HasBackendService check service exists by service id
func HasBackendService(serviceID string) bool {
	_, ok := servicesMap.Load(serviceID)
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
