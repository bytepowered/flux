package main

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/server"
	_ "github.com/bytepowered/flux/flux-node/transporter/dubbo"
	_ "github.com/bytepowered/flux/flux-node/transporter/echo"
	_ "github.com/bytepowered/flux/flux-node/transporter/http"
	_ "github.com/bytepowered/flux/flux-node/webecho"
)

import (
	_ "github.com/apache/dubbo-go/filter/filter_impl"
	_ "github.com/apache/dubbo-go/registry/zookeeper"
)

var (
	GitCommit string
	Version   string
	BuildDate string
)

// 注意：自定义实现main方法时，需要导入WebServer实现模块；
// 或者导入 _ "github.com/bytepowered/flux/webecho" 自动注册WebServer；
func main() {
	server.InitLogger()
	server.Bootstrap(flux.Build{CommitId: GitCommit, Version: Version, Date: BuildDate})
}
