package flux

import (
	"github.com/spf13/cast"
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestNamedValueSpec_ToString(t *testing.T) {
	cases := []TestCase{
		{
			Expected: "a",
			Actual:   NamedValueSpec{Value: []string{"a", "b"}}.GetString(),
			Message:  "should be [0] == a",
		},
		{
			Expected: "",
			Actual:   NamedValueSpec{}.GetString(),
			Message:  "default nil should be empty string",
		},
		{
			Expected: "",
			Actual:   NamedValueSpec{Value: []string{}}.GetString(),
			Message:  "empty array should be empty string",
		},
		{
			Expected: "",
			Actual:   NamedValueSpec{Value: nil}.GetString(),
			Message:  "nil value should be empty string",
		},
		{
			Expected: "0",
			Actual:   NamedValueSpec{Value: 0}.GetString(),
			Message:  "0 value should be '0' string",
		},
	}
	tester := assert2.New(t)
	for _, c := range cases {
		tester.Equal(c.Expected, c.Actual, c.Message)
	}
}

func TestParseEndpointModelV2(t *testing.T) {
	text := `{
    "endpointId": "auc:/api/users/register:1.0",
    "version": "1.0",
    "application": "auc",
    "httpPattern": "/api/users/register",
    "httpMethod": "POST",
	"serviceId": "com.foo.bar.testing.app.TestingAppService:testContext",
    "service": {
		"kind": "DubboService",
        "serviceId": "com.foo.bar.testing.app.TestingAppService:testContext",
        "interface": "com.foo.bar.testing.app.TestingAppService",
        "method": "testContext",
		"protocol": "DUBBO",
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
		"annotations": {
			"flux.go/rpc.group": "",
			"flux.go/rpc.version": "0.0.0.0.1"
		},
        "attributes": [
            {
                "name": "RpcProto",
                "value": "DUBBO"
            }
        ]
    },
    "permissions": [],
	"annotations": {
		"flux.go/static.model": true
	},
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
	AssertEndpointSpecWith(t, text, []EndpointSpecCase{
		{
			Expected: true,
			Actual:   func(endpoint *EndpointSpec) interface{} { return endpoint.IsValid() },
		},
		{
			Expected: "auc",
			Actual:   func(endpoint *EndpointSpec) interface{} { return endpoint.Application },
		},
		{
			Expected: "1.0",
			Actual:   func(endpoint *EndpointSpec) interface{} { return endpoint.Version },
		},
		{
			Expected: "/api/users/register",
			Actual:   func(endpoint *EndpointSpec) interface{} { return endpoint.HttpPattern },
		},
		{
			Expected: "POST",
			Actual:   func(endpoint *EndpointSpec) interface{} { return endpoint.HttpMethod },
		},
		{
			Expected: ":superadmin",
			Actual:   func(endpoint *EndpointSpec) interface{} { return endpoint.Attributes.Single("roles").GetString() },
		},
		{
			Expected: []string{"key=query:etag", "ttl=3600"},
			Actual: func(endpoint *EndpointSpec) interface{} {
				return endpoint.Attributes.Single("feature:cache").GetStrings()
			},
		},
		{
			Expected: true,
			Actual:   func(endpoint *EndpointSpec) interface{} { return cast.ToBool(endpoint.Attribute("Authorize").Value) },
		},
		{
			Expected: "DubboService",
			Actual:   func(endpoint *EndpointSpec) interface{} { return endpoint.Service.Kind },
		},
		{
			Expected: 3,
			Actual:   func(endpoint *EndpointSpec) interface{} { return len(endpoint.Service.Arguments) },
		},
		{
			Expected: "context",
			Actual:   func(endpoint *EndpointSpec) interface{} { return endpoint.Service.Arguments[0].Name },
		},
		{
			Expected: "COMPLEX",
			Actual:   func(endpoint *EndpointSpec) interface{} { return endpoint.Service.Arguments[0].StructType },
		},
		{
			Expected: "$body",
			Actual:   func(endpoint *EndpointSpec) interface{} { return endpoint.Service.Arguments[0].Fields[0].HttpName },
		},
		{
			Expected: "BODY",
			Actual:   func(endpoint *EndpointSpec) interface{} { return endpoint.Service.Arguments[0].Fields[0].HttpScope },
		},
		{
			Expected: "DUBBO",
			Actual:   func(endpoint *EndpointSpec) interface{} { return endpoint.Service.Protocol },
		},
		{
			Expected: "0.0.0.0.1",
			Actual: func(endpoint *EndpointSpec) interface{} {
				return endpoint.Service.Annotation(ServiceAnnotationRpcVersion).GetString()
			},
		},
	})
}

func AssertEndpointSpecWith(t *testing.T, text string, cases []EndpointSpecCase) {
	endpoint := EndpointSpec{}
	serializer := NewJsonSerializer()
	err := serializer.Unmarshal([]byte(text), &endpoint)
	tAssert := assert2.New(t)
	tAssert.Nil(err, "err must nil")
	for _, c := range cases {
		tAssert.Equal(c.Expected, c.Actual(&endpoint), c.Message)
	}
}

type EndpointSpecCase struct {
	Expected interface{}
	Actual   func(endpoint *EndpointSpec) interface{}
	Message  string
}
