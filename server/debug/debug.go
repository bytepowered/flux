package debug

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/webmidware"
	"github.com/labstack/gommon/random"
	https "net/http"
	_ "net/http/pprof"
)

const (
	HttpConfigDebugAuthUsername = "debug-auth-username"
	HttpConfigDebugAuthPassword = "debug-auth-password"
)

func BasicAuthMiddleware(config *flux.Configuration) flux.WebMiddleware {
	config.SetDefaults(map[string]interface{}{
		HttpConfigDebugAuthUsername: "fluxgo",
		HttpConfigDebugAuthPassword: random.String(8),
	})
	username := config.GetString(HttpConfigDebugAuthUsername)
	password := config.GetString(HttpConfigDebugAuthPassword)
	logger.Infow("Http debug feature: [ENABLED], Auth: BasicAuth", "username", username, "password", password)
	return webmidware.NewBasicAuthMiddleware(func(user string, pass string, webc flux.WebContext) (bool, error) {
		return user == username && pass == password, nil
	})
}

func QueryEndpointFeature(datamap map[string]*internal.MultiVersionEndpoint) flux.WebRouteHandler {
	// Endpoint查询
	json := ext.GetSerializer(ext.TypeNameSerializerJson)
	return func(webc flux.WebContext) error {
		if data, err := json.Marshal(queryEndpoints(datamap, webc)); nil != err {
			return err
		} else {
			return webc.ResponseWrite(https.StatusOK, flux.MIMEApplicationJSONCharsetUTF8, data)
		}
	}
}
