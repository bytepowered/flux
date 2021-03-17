package logger

import (
	"context"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	fluxpkg "github.com/bytepowered/flux/flux-pkg"
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
	fluxpkg.AssertNotNil(ctx, "<flux.context> must not nil in log trace")
	fields := map[string]string{
		"request-method": ctx.Method(),
		"request-uri":    ctx.URI(),
	}
	for k, v := range extras {
		fields[k] = v
	}
	endpoint := ctx.Endpoint()
	if nil != endpoint && endpoint.IsValid() {
		fields["appid"] = endpoint.Application
		fields["bizid"] = endpoint.GetAttr(flux.EndpointAttrTagBizId).GetString()
		fields["transporter-service"] = endpoint.Service.ServiceID()
		fields["transporter-permission"] = strings.Join(endpoint.PermissionServiceIds(), ",")
		fields["transporter-authorize"] = cast.ToString(endpoint.AttrAuthorize())
		fields["endpoint-version"] = endpoint.Version
		fields["endpoint-pattern"] = endpoint.HttpPattern
	}
	return TraceExtras(ctx.RequestId(), fields)
}

func TraceExtras(traceId string, extras map[string]string) flux.Logger {
	p := context.WithValue(context.Background(), TraceId, traceId)
	return ext.NewLoggerWith(context.WithValue(p, Extras, extras))
}
