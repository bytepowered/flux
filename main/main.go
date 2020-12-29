package main

import (
	"github.com/bytepowered/flux"
	_ "github.com/bytepowered/flux/backend/dubbo"
	_ "github.com/bytepowered/flux/backend/echo"
	_ "github.com/bytepowered/flux/backend/http"
	"github.com/bytepowered/flux/server"
	_ "github.com/bytepowered/flux/webecho"
)

var (
	GitCommit string
	Version   string
	BuildDate string
)

// 注意：自定义实现main方法时，需要导入WebServer实现模块；
// 或者导入 _ "github.com/bytepowered/flux/webecho" 自动注册WebServer；
func main() {
	server.InitDefaultLogger()
	server.Run(flux.BuildInfo{CommitId: GitCommit, Version: Version, Date: BuildDate})
}
