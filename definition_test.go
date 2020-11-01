package flux

import (
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestParseJsonTextToEndpoint(t *testing.T) {
	text := `{
    "version":"3.0",
    "application":"testapp",
		"authorize":true,
    "service": {
			"protocol":"DUBBO",
			"group":"myg",
			"version":"1.0.0",
			"retries":0,
			"interface":"foo.bar.Service",
			"method":"reportDetail",
			"arguments":[
							{
									"class":"java.lang.String",
									"generic":[],
									"name":"devId",
									"type":"PRIMITIVE",
									"httpName":"devId",
									"httpScope":"AUTO"
							},
							{
									"class":"java.lang.Integer",
									"generic":[],
									"name":"year",
									"type":"PRIMITIVE",
									"httpName":"year",
									"httpScope":"AUTO"
							},
							{
									"class":"java.lang.Integer",
									"generic":[],
									"name":"month",
									"type":"PRIMITIVE",
									"httpName":"month",
									"httpScope":"AUTO"
							},
							{
									"class":"java.lang.Integer",
									"generic":[],
									"name":"week",
									"type":"PRIMITIVE",
									"httpName":"week",
									"httpScope":"AUTO"
							}
					]
		},
    "httpPattern":"/api/foo/bar",
    "httpMethod":"POST",
		"extensions": {
				"key": "value",
				"bool": true
		},
		"permission": {
				"interface":"foo.bar.Service",
    		"method":"checkPermission",
				"protocol":"DUBBO",
				"arguments":[
						{
								"class":"java.lang.String",
								"generic":[],
								"name":"devId",
								"type":"PRIMITIVE",
								"httpName":"devId",
								"httpScope":"AUTO"
						},
						{
								"class":"java.lang.Integer",
								"generic":[],
								"name":"year",
								"type":"PRIMITIVE",
								"httpName":"year",
								"httpScope":"AUTO"
						}
				]
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
	assert.Equal(4, len(endpoint.Service.Arguments))
	assert.Equal("month", endpoint.Service.Arguments[2].Name)
	assert.Equal("month", endpoint.Service.Arguments[2].HttpName)
	assert.Equal("PRIMITIVE", endpoint.Service.Arguments[2].Type)

	assert.Equal("DUBBO", endpoint.Service.Protocol)
	assert.Equal("reportDetail", endpoint.Service.Method)

	assert.Equal(map[string]interface{}{"key": "value", "bool": true}, endpoint.Extensions)

	assert.Equal("checkPermission", endpoint.Permission.Method)
	assert.Equal(2, len(endpoint.Permission.Arguments))
	assert.Equal("year", endpoint.Permission.Arguments[1].Name)
	assert.Equal("year", endpoint.Permission.Arguments[1].HttpName)
	assert.Equal("PRIMITIVE", endpoint.Permission.Arguments[1].Type)
}
