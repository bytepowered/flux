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

func Trace(id string) flux.Logger {
	return ext.NewLoggerWith(context.WithValue(context.Background(), TraceId, id))
}

func TraceContext(ctx *flux.Context) flux.Logger {
	return TraceContextExtras(ctx, nil)
}

func TraceContextExtras(ctx *flux.Context, extras map[string]string) flux.Logger {
	flux.AssertNotNil(ctx, "<flux.context> must not nil in log trace")
	fields := map[string]string{
		"request.method": ctx.Method(),
		"request.uri":    ctx.URI(),
	}
	for k, v := range extras {
		fields[k] = v
	}
	endpoint := ctx.Endpoint()
	if nil != endpoint && endpoint.IsValid() {
		fields["endpoint.appid"] = endpoint.Application
		fields["endpoint.bizid"] = endpoint.Attributes.Single(flux.EndpointAttrTagBizId).ToString()
		fields["endpoint.version"] = endpoint.Version
		fields["endpoint.pattern"] = endpoint.HttpPattern
		fields["authorize"] = cast.ToString(endpoint.Authorize())
		fields["service"] = endpoint.Service.ServiceID()
		fields["service.permission"] = strings.Join(endpoint.PermissionIds(), ",")
	}
	return TraceExtras(ctx.RequestId(), fields)
}

func TraceExtras(traceId string, extras map[string]string) flux.Logger {
	p := context.WithValue(context.Background(), TraceId, traceId)
	return ext.NewLoggerWith(context.WithValue(p, Extras, extras))
}
