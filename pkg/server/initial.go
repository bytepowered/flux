package server

import (
	"context"
	"fmt"
	"github.com/bytepowered/fluxgo/pkg/common"
	"github.com/bytepowered/fluxgo/pkg/discovery"
	"github.com/bytepowered/fluxgo/pkg/ext"
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/logger"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	// Default logger factory
	ext.SetLoggerFactory(logger.DefaultFactory)
	// 参数查找与解析函数
	ext.SetLookupScopedValueFunc(common.LookupValueByScoped)
	// Serializer
	// Default: JSON
	serializer := flux.NewJsonSerializer()
	ext.RegisterSerializer(ext.TypeNameSerializerDefault, serializer)
	ext.RegisterSerializer(ext.TypeNameSerializerJson, serializer)
	// Endpoint discovery
	ext.RegisterMetadataDiscovery(discovery.NewZookeeperMetadataDiscovery(discovery.ZookeeperId))
	ext.RegisterMetadataDiscovery(discovery.NewResourceMetadataDiscovery(discovery.ResourceId))
}

// InitConfig 初始化配置
func InitConfig(file string) error {
	viper.SetConfigName(file)
	viper.AddConfigPath("./")
	viper.AddConfigPath("./conf.d")
	viper.AddConfigPath("/etc/flux/conf.d")
	logger.Infof("Using config, file: %s", file)
	if err := viper.ReadInConfig(); nil != err {
		return fmt.Errorf("read config to viper, file: %s, error: %w", file, err)
	}
	return nil
}

// InitLogger 初始化日志
func InitLogger(file string) error {
	config, err := logger.LoadConfig(file)
	if nil != err {
		panic(err)
	}
	sugar := logger.NewZapLogger(config)
	logger.SetSimpleLogger(sugar)
	zap.ReplaceGlobals(sugar.Desugar())
	ext.SetLoggerFactory(func(values context.Context) flux.Logger {
		if nil == values {
			return sugar
		}
		if traceId := values.Value(logger.TraceId); nil != traceId {
			return sugar.With(zap.String(logger.TraceId, cast.ToString(traceId)))
		}
		return sugar
	})
	return nil
}
