package fluxinspect

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/common"
	"strings"
)

func queryWithEndpointFilters(data map[string]*flux.MVCEndpoint, filters ...EndpointFilter) []map[string]*flux.Endpoint {
	items := make([]map[string]*flux.Endpoint, 0, 16)
DataLoop:
	for _, v := range data {
		for _, filter := range filters {
			// 任意Filter返回True
			if filter(v) {
				items = append(items, v.ToSerializable())
				continue DataLoop
			}
		}
	}
	return items
}

func queryMatch(input, expected string) bool {
	input, expected = strings.ToLower(input), strings.ToLower(expected)
	return strings.Contains(expected, input)
}

func send(webex flux.ServerWebContext, status int, payload interface{}) error {
	bytes, err := common.SerializeObject(payload)
	if nil != err {
		return err
	}
	return webex.Write(status, flux.MIMEApplicationJSONCharsetUTF8, bytes)
}
