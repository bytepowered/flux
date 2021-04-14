package fluxinspect

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
)

const (
	queryKeyServiceId0 = "service-id"
	queryKeyServiceId1 = "service"
)

var (
	serviceQueryKeys = []string{queryKeyServiceId0, queryKeyServiceId1}
)

func ServicesHandler(ctx flux.ServerWebContext) error {
	for _, key := range serviceQueryKeys {
		if id := ctx.QueryVar(key); "" != id {
			service, ok := ext.TransporterServiceById(id)
			if ok {
				return send(ctx, flux.StatusOK, service)
			} else {
				return send(ctx, flux.StatusNotFound, map[string]string{
					"status":     "failed",
					"message":    "service not found",
					"service-id": id,
				})
			}
		}
	}
	return send(ctx, flux.StatusBadRequest, map[string]string{
		"status":  "failed",
		"message": "param is required: serviceId",
	})
}
