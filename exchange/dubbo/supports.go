package dubbo

import (
	dubbogo "github.com/apache/dubbo-go/config"
	"github.com/apache/dubbo-go/protocol/dubbo"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"time"
)

func assemble(arguments []flux.Argument) (types []string, values []interface{}) {
	size := len(arguments)
	types = make([]string, size)
	values = make([]interface{}, size)
	for i, arg := range arguments {
		types[i] = arg.TypeClass
		if flux.ArgumentTypePrimitive == arg.ArgType {
			values[i] = arg.ArgValue.Value()
		} else if flux.ArgumentTypeComplex == arg.ArgType {
			values[i] = argToMap(arg)
		} else {
			logger.Warnf("Unsupported parameter type: %s", arg.ArgType)
		}
	}
	return
}

func argToMap(arg flux.Argument) map[string]interface{} {
	m := make(map[string]interface{}, 1+len(arg.Fields))
	m["class"] = arg.TypeClass
	for _, field := range arg.Fields {
		if flux.ArgumentTypePrimitive == field.ArgType {
			m[field.ArgName] = field.ArgValue.Value()
		} else if flux.ArgumentTypeComplex == arg.ArgType {
			m[field.ArgName] = argToMap(field)
		} else {
			logger.Warnf("Unsupported parameter type: %s", arg.ArgType)
		}
	}
	return m
}

func newReference(endpoint *flux.Endpoint, config flux.Config) *dubbogo.ReferenceConfig {
	ifaceName := endpoint.UpstreamUri
	timeout := getConfig(endpoint.RpcTimeout, "timeout", config, "3000")
	retries := getConfig(endpoint.RpcRetries, "retries", config, "0")
	reference := &dubbogo.ReferenceConfig{
		InterfaceName:  ifaceName,
		Version:        endpoint.RpcVersion,
		Group:          endpoint.RpcGroup,
		RequestTimeout: timeout,
		Cluster:        config.StringOrDefault("cluster", "failover"),
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

func getConfig(specifiedValue string, configKey string, configs flux.Config, defValue string) string {
	if "" != specifiedValue {
		return specifiedValue
	}
	return configs.StringOrDefault(configKey, defValue)
}
