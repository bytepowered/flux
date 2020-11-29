package ext

import (
	"fmt"
	"github.com/bytepowered/flux"
	"sync"
)

var (
	_serviceNotFound flux.BackendService
	_servicesMap     *sync.Map
)

func StoreBackendService(service flux.BackendService) {
	id := _ensureServiceId(&service)
	_servicesMap.Store(id, service)
}

func LoadBackendService(serviceId string) (flux.BackendService, bool) {
	v, ok := _servicesMap.Load(serviceId)
	if ok {
		return v.(flux.BackendService), true
	}
	return _serviceNotFound, false
}

func RemoveBackendService(serviceId string) {
	_servicesMap.Delete(serviceId)
}

func HasBackendService(serviceId string) bool {
	_, ok := _servicesMap.Load(serviceId)
	return ok
}

func _ensureServiceId(service *flux.BackendService) string {
	id := service.ServiceId
	if "" == id {
		id = service.Interface + ":" + service.Method
	}
	if len(id) < len("a:b") {
		panic(fmt.Sprintf("BackendService must has an Id, service: %+v", service))
	}
	return id
}
