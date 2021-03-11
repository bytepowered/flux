package logger

import (
	"context"
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"strings"
)

const (
	TraceId = "trace-id"
	Extras  = "extras"
)

func Trace(traceId string) flux2.Logger {
	return ext.NewLoggerWith(context.WithValue(context.Background(), TraceId, traceId))
}

func TraceExtras(traceId string, extras map[string]string) flux2.Logger {
	p := context.WithValue(context.Background(), TraceId, traceId)
	return ext.NewLoggerWith(context.WithValue(p, Extras, extras))
}

func TraceContext(ctx flux2.Context) flux2.Logger {
	return TraceContextExtras(ctx, nil)
}

func TraceContextExtras(ctx flux2.Context, extras map[string]string) flux2.Logger {
	if nil == ctx {
		return Trace("no-trace-id")
	}
	logger := ctx.Logger()
	if logger == nil {
		logger = zap.S()
	}
	fields := map[string]string{
		"request-id":     ctx.RequestId(),
		"request-method": ctx.Method(),
		"request-uri":    ctx.URI(),
	}
	for k, v := range extras {
		fields[k] = v
	}
	endpoint := ctx.Endpoint()
	if endpoint.IsValid() {
		fields["appid"] = endpoint.Application
		fields["bizid"] = endpoint.GetAttr(flux2.EndpointAttrTagBizId).GetString()
		fields["backend-service"] = endpoint.Service.ServiceID()
		fields["backend-permission"] = strings.Join(endpoint.PermissionServiceIds(), ",")
		fields["backend-authorize"] = cast.ToString(endpoint.AttrAuthorize())
		fields["endpoint-version"] = endpoint.Version
		fields["endpoint-pattern"] = endpoint.HttpPattern
		return TraceExtras(ctx.RequestId(), fields)
	} else {
		return Trace(ctx.RequestId())
	}
}
