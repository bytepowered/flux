package main

import (
	"github.com/bytepowered/flux/flux-node"
	_ "github.com/bytepowered/flux/flux-node/backend/dubbo"
	_ "github.com/bytepowered/flux/flux-node/backend/echo"
	_ "github.com/bytepowered/flux/flux-node/backend/http"
	"github.com/bytepowered/flux/flux-node/boot"
	_ "github.com/bytepowered/flux/flux-node/webserver"
)

import (
	_ "github.com/apache/dubbo-go/cluster/cluster_impl"
	_ "github.com/apache/dubbo-go/cluster/loadbalance"
	_ "github.com/apache/dubbo-go/filter/filter_impl"
	_ "github.com/apache/dubbo-go/registry/protocol"
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
	boot.InitLogger()
	boot.Bootstrap(flux.Build{CommitId: GitCommit, Version: Version, Date: BuildDate})
}
