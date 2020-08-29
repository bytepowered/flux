package server

import (
	"github.com/bytepowered/flux"
)

const (
	HttpConfigDebugAuthUsername = "debug-auth-username"
	HttpConfigDebugAuthPassword = "debug-auth-password"
)

func (s *HttpServer) debugFeatures(config *flux.Configuration) {
	/*config.SetDefaults(map[string]interface{}{
		HttpConfigDebugAuthUsername: "fluxgo",
		HttpConfigDebugAuthPassword: random.String(8),
	})
	username := config.GetString(HttpConfigDebugAuthUsername)
	password := config.GetString(HttpConfigDebugAuthPassword)
	logger.Infow("Http debug feature: [ENABLED], Auth: BasicAuth", "username", username, "password", password)
	auth := webmidware.NewBasicAuthMiddleware(func(user string, pass string, webc webx.WebContext) (bool, error) {
		return user == username && pass == password, nil
	})
	// Debug查询接口
	debugHandler := https.DefaultServeMux
	s.AddStdHttpHandler("GET", DebugPathVars, debugHandler, auth)
	s.AddStdHttpHandler("GET", DebugPathPprof, debugHandler, auth)
	// Endpoint查询
	json := ext.GetSerializer(ext.TypeNameSerializerJson)
	s.AddStdHttpHandler("GET", DebugPathEndpoints, func(webc webx.WebContext) error {
		if data, err := json.Marshal(queryEndpoints(s.mvEndpointMap, webc)); nil != err {
			return err
		} else {
			webc.SetResponseHeader(webx.HeaderContentType, webx.MIMEApplicationJSONCharsetUTF8)
			return webc.ResponseWrite(https.StatusOK, data)
		}
	}, auth)*/
}
