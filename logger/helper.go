package logger

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"strings"
)

const (
	TraceId = "trace-id"
	Extras  = "extras"
)

func Trace(traceId string) flux.Logger {
	return ext.NewLoggerWith(context.WithValue(context.Background(), TraceId, traceId))
}

func TraceExtras(traceId string, extras map[string]string) flux.Logger {
	p := context.WithValue(context.Background(), TraceId, traceId)
	return ext.NewLoggerWith(context.WithValue(p, Extras, extras))
}

func TraceContext(ctx flux.Context) flux.Logger {
	return TraceContextExtras(ctx, nil)
}

func TraceContextExtras(ctx flux.Context, extras map[string]string) flux.Logger {
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
		fields["bizid"] = endpoint.GetAttr(flux.EndpointAttrTagBizId).GetString()
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
