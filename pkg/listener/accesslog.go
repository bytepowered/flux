package listener

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/logger"
	"time"
)

func NewAccessLogFilter() flux.WebFilter {
	return func(next flux.WebHandlerFunc) flux.WebHandlerFunc {
		return func(webc flux.WebContext) error {
			multiWriter := NewEmptyHttpResponseMultiWriter(webc.ResponseWriter())
			webc.SetResponseWriter(multiWriter)
			defer func(trace flux.Logger, start time.Time) {
				trace.Infow("SERVER:TRAFFIC:LOG",
					"listener-id", webc.WebListener().ListenerId(),
					"remote-ip", webc.RemoteAddr(),
					"request.uri", webc.URI(),
					"request.method", webc.Method(),
					"request.header", webc.HeaderVars(),
					"response.header", webc.ResponseWriter().Header(),
					"response.status", multiWriter.StatusCode(),
					"latency", time.Since(start).String(),
				)
			}(logger.Trace(webc.RequestId()), time.Now())
			return next(webc)
		}
	}
}
