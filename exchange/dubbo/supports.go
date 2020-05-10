package dubbo

import (
	dubbogo "github.com/apache/dubbo-go/config"
	"github.com/apache/dubbo-go/protocol/dubbo"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"time"
)

func assemble(arguments []flux.Argument) (types []string, values []interface{}) {
	size := len(arguments)
	types = make([]string, size)
	values = make([]interface{}, size)
	for i, arg := range arguments {
		types[i] = arg.TypeClass
		if flux.ArgumentTypePrimitive == arg.Type {
			values[i] = arg.HttpValue.Value()
		} else if flux.ArgumentTypeComplex == arg.Type {
			values[i] = argToMap(arg)
		} else {
			logger.Warnf("Unsupported parameter type: %s", arg.Type)
		}
	}
	return
}

func argToMap(arg flux.Argument) map[string]interface{} {
	m := make(map[string]interface{}, 1+len(arg.Fields))
	m["class"] = arg.TypeClass
	for _, field := range arg.Fields {
		if flux.ArgumentTypePrimitive == field.Type {
			m[field.Name] = field.HttpValue.Value()
		} else if flux.ArgumentTypeComplex == arg.Type {
			m[field.Name] = argToMap(field)
		} else {
			logger.Warnf("Unsupported parameter type: %s", arg.Type)
		}
	}
	return m
}

func newReference(endpoint *flux.Endpoint, config pkg.Configuration) *dubbogo.ReferenceConfig {
	ifaceName := endpoint.UpstreamUri
	timeout := getConfig(endpoint.RpcTimeout, "timeout", config, "3000")
	retries := getConfig(endpoint.RpcRetries, "retries", config, "0")
	reference := &dubbogo.ReferenceConfig{
		InterfaceName:  ifaceName,
		Version:        endpoint.RpcVersion,
		Group:          endpoint.RpcGroup,
		RequestTimeout: timeout,
		Cluster:        config.GetStringOr("cluster", "failover"),
		Retries:        retries,
		Protocol:       dubbo.DUBBO,
		Generic:        true,
	}
	logger.Infof("Create dubbo reference-config, iface: %s, config: %+v", ifaceName, config)
	reference.GenericLoad(ifaceName)
	<-time.After(time.Millisecond * 100)
	logger.Infof("Create dubbo reference-config, iface: %s, OK", ifaceName)
	return reference
}

func getConfig(specifiedValue string, configKey string, config pkg.Configuration, defValue string) string {
	if "" != specifiedValue {
		return specifiedValue
	}
	return config.GetStringOr(configKey, defValue)
}
