package ext

import (
	"fmt"
	"github.com/bytepowered/flux/flux-node"
	"sync"
)

var (
	serviceNotFound flux.TransporterService
	servicesMap     *sync.Map = new(sync.Map)
)

func RegisterTransporterServiceById(id string, service flux.TransporterService) {
	servicesMap.Store(id, service)
}

// RegisterTransporterService store transporter service
func RegisterTransporterService(service flux.TransporterService) {
	id := _ensureServiceID(&service)
	RegisterTransporterServiceById(id, service)
}

func TransporterServices() map[string]flux.TransporterService {
	out := make(map[string]flux.TransporterService, 512)
	endpoints.Range(func(key, value interface{}) bool {
		out[key.(string)] = value.(flux.TransporterService)
		return true
	})
	return out
}

// TransporterServiceById load transporter service by serviceId
func TransporterServiceById(serviceID string) (flux.TransporterService, bool) {
	v, ok := servicesMap.Load(serviceID)
	if ok {
		return v.(flux.TransporterService), true
	}
	return serviceNotFound, false
}

// RemoveTransporterService remove transporter service by serviceId
func RemoveTransporterService(serviceID string) {
	servicesMap.Delete(serviceID)
}

// HasTransporterService check service exists by service id
func HasTransporterService(serviceID string) bool {
	_, ok := servicesMap.Load(serviceID)
	return ok
}

func _ensureServiceID(service *flux.TransporterService) string {
	id := service.ServiceId
	if id == "" {
		id = service.Interface + ":" + service.Method
	}
	if len(id) < len("a:b") {
		panic(fmt.Sprintf("Transporter must has an Id, service: %+v", service))
	}
	return id
}
