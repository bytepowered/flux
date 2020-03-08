package main

import (
	"github.com/bytepowered/flux/bootstrap"
	"github.com/bytepowered/flux/internal"
)

var (
	GitCommit string
	Version   string
	BuildDate string
)

func main() {
	bootstrap.Run(internal.BuildVersion{CommitId: GitCommit, Version: Version, Date: BuildDate})
}
