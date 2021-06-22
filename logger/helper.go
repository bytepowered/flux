package logger

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
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
	if nil != endpoint && endpoint.Valid() {
		fields["endpoint.appid"] = endpoint.Application
		fields["endpoint.bizkey"] = endpoint.Annotation(flux.EndpointAnnoNameBizKey).ToString()
		fields["endpoint.version"] = endpoint.Version
		fields["endpoint.pattern"] = endpoint.HttpPattern
		fields["endpoint.authorize"] = endpoint.Annotation(flux.EndpointAnnoNameAuthorize).ToString()
		fields["endpoint.service"] = endpoint.Service.ServiceID()
		fields["endpoint.permission"] = strings.Join(endpoint.Attributes.Multiple(flux.EndpointAttrTagPermission).Strings(), ",")
	}
	return TraceExtras(ctx.RequestId(), fields)
}

func TraceExtras(traceId string, extras map[string]string) flux.Logger {
	p := context.WithValue(context.Background(), TraceId, traceId)
	return ext.NewLoggerWith(context.WithValue(p, Extras, extras))
}
