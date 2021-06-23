package server

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
)

func IsDisabled(config *flux.Configuration) bool {
	return config.GetBool("disable") || config.GetBool("disabled")
}

func DefaultRequestVersionLocateFunc(webex flux.WebContext) (version string) {
	return webex.HeaderVar(DefaultHttpHeaderVersion)
}
