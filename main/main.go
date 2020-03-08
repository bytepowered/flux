package main

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/server"
)

var (
	GitCommit string
	Version   string
	BuildDate string
)

func main() {
	server.Run(flux.BuildInfo{CommitId: GitCommit, Version: Version, Date: BuildDate})
}
