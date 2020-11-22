package webhandler

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/support"
	"net/http"
)

var (
	HealthStateCmdNotFound     = []byte(`{"status":"error", "message": "COMMAND_NOT_FOUND"}`)
	HealthStateCmdNotSupported = []byte(`{"status":"error", "message": "COMMAND_NOT_SUPPORTED"}`)
)

// HealthStateFunc 应用健康状态检查函数
type HealthStateFunc func(webc flux.WebContext) (statusCode int, message []byte)

// HealthCheckConfig 应用健康状态检查配置
type HealthCheckConfig struct {
	CommandLookupKey string
	CommandHandlers  map[string]HealthStateFunc
}

// NewHealthCheckWebRouteHandlerFactory 根据配置构建应用健康检查WebHandler
func NewHealthCheckWebRouteHandlerFactory(config HealthCheckConfig) flux.WebHandler {
	if config.CommandLookupKey == "" {
		logger.Panicw("Health check config, requires: CommandLookupKey")
	}
	if config.CommandHandlers == nil {
		logger.Panicw("Health check config, requires: CommandHandlers")
	}
	return func(webc flux.WebContext) error {
		cmd := support.LookupWebContextByExpr(config.CommandLookupKey, webc)
		if cmd == "" {
			return webc.Write(http.StatusBadRequest, flux.MIMEApplicationJSONCharsetUTF8, HealthStateCmdNotFound)
		}
		handler, ok := config.CommandHandlers[cmd]
		if !ok {
			return webc.Write(http.StatusBadRequest, flux.MIMEApplicationJSONCharsetUTF8, HealthStateCmdNotSupported)
		}
		status, data := handler(webc)
		return webc.Write(status, flux.MIMEApplicationJSONCharsetUTF8, data)
	}
}
