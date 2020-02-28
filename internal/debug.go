package internal

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/extension"
	"github.com/bytepowered/flux/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/random"
	"net/http"
)

// Enable debug features
func (a *Application) enabledDebugFeatures(httpConfig flux.Config) {
	basicAuthC := httpConfig.Config("BasicAuth")
	username := basicAuthC.StringOrDefault("username", "flux")
	password := basicAuthC.StringOrDefault("password", random.String(8))
	logger.Infof("Http debug feature: <Enabled>, basic-auth: username=%s, password=%s", username, password)
	authMiddleware := middleware.BasicAuth(func(u string, p string, c echo.Context) (bool, error) {
		return u == username && p == password, nil
	})
	debugHandler := echo.WrapHandler(http.DefaultServeMux)
	a.httpServer.GET("/debug/vars", debugHandler, authMiddleware)
	a.httpServer.GET("/debug/pprof/*", debugHandler, authMiddleware)
	a.httpServer.GET("/debug/endpoints", func(c echo.Context) error {
		m := make(map[string]interface{})
		for k, v := range a.endpointMvMap {
			m[k] = v.ToSerializableMap()
		}
		serializer := extension.GetSerializer(extension.TypeNameSerializerJson)
		if data, err := serializer.Marshal(m); nil != err {
			return err
		} else {
			return c.JSONBlob(200, data)
		}
	}, authMiddleware)
}
