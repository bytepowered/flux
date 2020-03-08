package main

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/bootstrap"
)

var (
	GitCommit string
	Version   string
	BuildDate string
)

func main() {
	bootstrap.Run(flux.BuildInfo{CommitId: GitCommit, Version: Version, Date: BuildDate})
}
