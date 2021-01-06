package logger

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/spf13/cast"
	"strings"
)

const (
	TraceId = "trace-id"
	Extras  = "extras"
)

func Trace(traceId string) flux.Logger {
	return With(context.WithValue(context.Background(), TraceId, traceId))
}

func TraceWith(traceId string, fields map[string]string) flux.Logger {
	p := context.WithValue(context.Background(), TraceId, traceId)
	return With(context.WithValue(p, Extras, fields))
}

func With(values context.Context) flux.Logger {
	return ext.NewLoggerWith(values)
}

func TraceContext(ctx flux.Context) flux.Logger {
	return TraceContextWith(ctx, nil)
}

func TraceContextWith(ctx flux.Context, extraFields map[string]string) flux.Logger {
	if nil == ctx {
		return Trace("no-trace-id")
	}
	if ctxLogger, ok := ctx.GetContextLogger(); ok {
		return ctxLogger
	}
	fields := map[string]string{
		"request-id":     ctx.RequestId(),
		"request-method": ctx.Method(),
		"request-uri":    ctx.RequestURI(),
	}
	for k, v := range extraFields {
		fields[k] = v
	}
	endpoint := ctx.Endpoint()
	if endpoint.IsValid() {
		fields["backend-appid"] = endpoint.Application
		fields["backend-service"] = endpoint.Service.ServiceID()
		fields["backend-permission"] = strings.Join(endpoint.PermissionServiceIds(), ",")
		fields["backend-authorize"] = cast.ToString(endpoint.AttrAuthorize())
		fields["endpoint-version"] = endpoint.Version
		fields["endpoint-pattern"] = endpoint.HttpPattern
		return TraceWith(ctx.RequestId(), fields)
	} else {
		return Trace(ctx.RequestId())
	}
}
