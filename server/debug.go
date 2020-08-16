package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/webx"
	"github.com/bytepowered/flux/webx/middleware"
	"github.com/labstack/gommon/random"
	https "net/http"
)

const (
	HttpConfigDebugAuthUsername = "debug-auth-username"
	HttpConfigDebugAuthPassword = "debug-auth-password"
)

func (s *HttpServer) debugFeatures(config *flux.Configuration) {
	config.SetDefaults(map[string]interface{}{
		HttpConfigDebugAuthUsername: "fluxgo",
		HttpConfigDebugAuthPassword: random.String(8),
	})
	username := config.GetString(HttpConfigDebugAuthUsername)
	password := config.GetString(HttpConfigDebugAuthPassword)
	logger.Infow("Http debug feature: [ENABLED], Auth: BasicAuth", "username", username, "password", password)
	auth := middleware.NewBasicAuthMiddleware(func(user string, pass string, webc webx.WebContext) (bool, error) {
		return user == username && pass == password, nil
	})
	// Debug查询接口
	debugHandler := webx.AdaptHttpHandler(https.DefaultServeMux)
	s.AddHttpHandler("GET", DebugPathVars, debugHandler, auth)
	s.AddHttpHandler("GET", DebugPathPprof, debugHandler, auth)
	// Endpoint查询
	json := ext.GetSerializer(ext.TypeNameSerializerJson)
	s.AddHttpHandler("GET", DebugPathEndpoints, func(webc webx.WebContext) error {
		if data, err := json.Marshal(queryEndpoints(s.mvEndpointMap, webc)); nil != err {
			return err
		} else {
			webc.ResponseHeader().Set(webx.HeaderContentType, webx.MIMEApplicationJSONCharsetUTF8)
			return webc.ResponseWrite(https.StatusOK, data)
		}
	}, auth)
}
