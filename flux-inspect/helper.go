package fluxinspect

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"strings"
)

func queryMatch(input, expected string) bool {
	input, expected = strings.ToLower(input), strings.ToLower(expected)
	return strings.Contains(expected, input)
}

func send(webex flux.ServerWebContext, status int, payload interface{}) error {
	bytes, err := ext.JSONMarshalObject(payload)
	if nil != err {
		return err
	}
	return webex.Write(status, flux.MIMEApplicationJSONCharsetUTF8, bytes)
}
