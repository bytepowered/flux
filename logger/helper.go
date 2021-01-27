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

func With(traceId string) flux.Logger {
	return ext.NewLoggerWith(context.WithValue(context.Background(), TraceId, traceId))
}

func WithFields(traceId string, fields map[string]string) flux.Logger {
	p := context.WithValue(context.Background(), TraceId, traceId)
	return ext.NewLoggerWith(context.WithValue(p, Extras, fields))
}

func WithContext(ctx flux.Context) flux.Logger {
	return WithContextExtras(ctx, nil)
}

func WithContextExtras(ctx flux.Context, extraFields map[string]string) flux.Logger {
	if nil == ctx {
		return With("no-trace-id")
	}
	logger := ctx.GetLogger()
	if logger == nil {
		logger = zap.S()
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
		return WithFields(ctx.RequestId(), fields)
	} else {
		return With(ctx.RequestId())
	}
}
