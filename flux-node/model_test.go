package flux

import (
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestParseJsonTextToEndpoint(t *testing.T) {
	text := `{
    "endpointId": "auc:/api/users/register:1.0",
    "version": "1.0",
    "application": "auc",
    "httpPattern": "/api/users/register",
    "httpMethod": "POST",
    "service": {
        "serviceId": "com.foo.bar.testing.app.TestingAppService:testContext",
        "interface": "com.foo.bar.testing.app.TestingAppService",
        "method": "testContext",
        "arguments": [
            {
                "class": "com.foo.bar.endpoint.support.BizAppContext",
                "generic": [],
                "name": "context",
                "type": "COMPLEX",
                "fields": [
                    {
                        "class": "java.lang.String",
                        "generic": [],
                        "name": "body",
                        "type": "PRIMITIVE",
                        "httpName": "$body",
                        "httpScope": "BODY"
                    }
                ]
            },
            {
                "class": "java.util.Map",
                "generic": [
                    "java.lang.String",
                    "java.lang.String"
                ],
                "name": "attrs",
                "type": "PRIMITIVE",
                "httpName": "$attrs",
                "httpScope": "ATTRS"
            },
            {
                "class": "java.util.List",
                "generic": [
                    "java.lang.String"
                ],
                "name": "roles",
                "type": "PRIMITIVE",
                "httpName": "roles",
                "httpScope": "ATTR"
            }
        ],
        "attributes": [
            {
                "name": "RpcProto",
                "value": "DUBBO"
            },
            {
                "name": "RpcGroup",
                "value": ""
            },
            {
                "name": "RpcVersion",
                "value": ""
            }
        ]
    },
    "permissions": [],
    "attributes": [
        {
            "name": "feature:cache",
            "value": [
                "key=query:etag",
                "ttl=3600"
            ]
        },
        {
            "name": "roles",
            "value": [
                ":superadmin"
            ]
        },
        {
            "name": "Authorize",
            "value": true
        },
        {
            "name": "protoType",
            "value": "Gateway"
        }
    ]
}`
	endpoint := Endpoint{}
	serializer := NewJsonSerializer()
	err := serializer.Unmarshal([]byte(text), &endpoint)
	assert := assert2.New(t)
	assert.NoError(err, "Should not error")
	assert.True(endpoint.IsValid())
	assert.Equal("auc", endpoint.Application)
	assert.Equal("1.0", endpoint.Version)
	assert.Equal("/api/users/register", endpoint.HttpPattern)
	assert.Equal("POST", endpoint.HttpMethod)
	assert.False(endpoint.Permission.IsValid())
	assert.True(len(endpoint.Permissions) == 0)
	assert.Equal(":superadmin", endpoint.GetAttr("roles").GetString())
	assert.Equal([]string{"key=query:etag", "ttl=3600"}, endpoint.GetAttr("feature:cache").GetStringSlice())
	assert.Equal(true, endpoint.GetAttr("Authorize").GetBool())
	assert.Equal(3, len(endpoint.Service.Arguments))
	assert.Equal("context", endpoint.Service.Arguments[0].Name)
	assert.Equal("COMPLEX", endpoint.Service.Arguments[0].Type)
	assert.Equal("$body", endpoint.Service.Arguments[0].Fields[0].HttpName)
	assert.Equal("BODY", endpoint.Service.Arguments[0].Fields[0].HttpScope)
	assert.Equal("DUBBO", endpoint.Service.RpcProto())
}
