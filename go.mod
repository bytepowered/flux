module github.com/bytepowered/flux

go 1.14

require (
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/apache/dubbo-go v1.5.6
	github.com/apache/dubbo-go-hessian2 v1.9.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dubbogo/go-zookeeper v1.0.3
	github.com/jinzhu/copier v0.3.2
	github.com/json-iterator/go v1.1.10
	github.com/labstack/echo/v4 v4.1.16
	github.com/labstack/gommon v0.3.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/spf13/cast v1.3.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.16.0
	golang.org/x/net v0.0.0-20201021035429-f5854403a974
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/apache/dubbo-go => github.com/yongjiapro/dubbo-go v1.5.6-rc3
