# Flux : MicroService Gateway for Dubbo/gRpc/Http

Flux Gateway 是一个基于 Golang 原生开发的微服务网关，支持Dubbo、HTTP以及gRPC等协议。

## Config

运行时须要指定环境变量：

> Note: Dubbogo强制地自动读取配置，过程不受管控，必须指定以下环境变量值。

```bash
export APP_LOG_CONF_FILE=./conf.d/log.yml
export CONF_CONSUMER_FILE_PATH=./conf.d/dubbo.yml
export CONF_PROVIDER_FILE_PATH=./conf.d/dubbo.yml
```

IDE运行

> Note: Dubbogo强制地自动读取配置，过程不受管控，必须指定以下环境变量值。

`
APP_LOG_CONF_FILE=./conf.d/log.yml;CONF_CONSUMER_FILE_PATH=./conf.d/dubbo.yml;CONF_PROVIDER_FILE_PATH=./conf.d/dubbo.yml
`

## LICENSE

Fluxgo is licensed under the Mulan PSL v2 license. Check the [LICENSE_CN](LICENSE) / [LICENSE_EN](LICENSE_EN) file for details.

### License report

Check the [LICENSES REPORT](LICENSES-REPORT.txt) file for details.
