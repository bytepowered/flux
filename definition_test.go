package flux

import (
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestParseJsonTextToEndpoint(t *testing.T) {
	text := `{
    "version":"3.0",
    "application":"testapp",
    "protocol":"DUBBO",
    "rpcGroup":"myg",
    "rpcVersion":"1.0.0",
    "rpcRetries":0,
    "authorize":true,
    "upstreamUri":"foo.bar.Service",
    "upstreamMethod":"reportDetail",
    "httpPattern":"/api/foo/bar",
    "httpMethod":"POST",
    "arguments":[
        {
            "typeClass":"java.lang.String",
            "typeGeneric":[],
            "argName":"devId",
            "argType":"PRIMITIVE",
            "httpName":"devId",
            "httpScope":"AUTO"
        },
        {
            "typeClass":"java.lang.Integer",
            "typeGeneric":[],
            "argName":"year",
            "argType":"PRIMITIVE",
            "httpName":"year",
            "httpScope":"AUTO"
        },
        {
            "typeClass":"java.lang.Integer",
            "typeGeneric":[],
            "argName":"month",
            "argType":"PRIMITIVE",
            "httpName":"month",
            "httpScope":"AUTO"
        },
        {
            "typeClass":"java.lang.Integer",
            "typeGeneric":[],
            "argName":"week",
            "argType":"PRIMITIVE",
            "httpName":"week",
            "httpScope":"AUTO"
        }
    ],
		"extensions": {
				"key": "value",
				"bool": true
		},
		"f0": "bar",
		"f1": "bar"
}`
	endpoint := Endpoint{}
	serializer := NewJsonSerializer()
	err := serializer.Unmarshal([]byte(text), &endpoint)
	assert := assert2.New(t)
	assert.NoError(err, "Should not error")
	assert.Equal("/api/foo/bar", endpoint.HttpPattern)
	assert.Equal(4, len(endpoint.Arguments))
	assert.Equal("month", endpoint.Arguments[2].Name)
	assert.Equal("month", endpoint.Arguments[2].HttpName)
	assert.Equal("PRIMITIVE", endpoint.Arguments[2].Type)

	assert.Equal(map[string]interface{}{"key": "value", "bool": true}, endpoint.Extensions)
}
