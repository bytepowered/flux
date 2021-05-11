package discovery

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/remoting"
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestParseEndpointModelV1(t *testing.T) {
	text := `{
    "application": "myapp",
    "authorize": true,
    "endpointId": "myapp:/v1/api/seller.adm.activity.list:1.0",
    "extensions": {
        "protoType": "GatewayAtop"
    },
    "httpMethod": "POST",
    "httpPattern": "/v1/api/seller.adm.activity.list",
    "permissions": [],
    "service": {
        "interface": "com.foo.bar.act.IActivityService",
        "method": "list",
        "rpcGroup": "",
        "rpcProto": "DUBBO",
        "rpcVersion": "",
        "serviceId": "com.foo.bar.act.IActivityService:list"
    },
    "serviceId": "com.foo.bar.act.IActivityService:list",
    "version": "1.0"
}`
	serializer := flux.NewJsonSerializer()
	ext.RegisterSerializer(ext.TypeNameSerializerJson, serializer)
	evt, err := NewEndpointEvent([]byte(text), remoting.EventTypeNodeAdd)
	if nil != err {
		t.Fatal("parse endpoint event failed, err :", err)
	}
	//flux.Assert
	AssertEndpointModel(t, &evt.Endpoint, []AssertCase{
		{
			Expected: true,
			Actual:   func(endpoint *flux.Endpoint) interface{} { return endpoint.IsValid() },
		},
		{
			Expected: "myapp",
			Actual:   func(endpoint *flux.Endpoint) interface{} { return endpoint.Application },
		},
		{
			Expected: "1.0",
			Actual:   func(endpoint *flux.Endpoint) interface{} { return endpoint.Version },
		},
		{
			Expected: "/v1/api/seller.adm.activity.list",
			Actual:   func(endpoint *flux.Endpoint) interface{} { return endpoint.HttpPattern },
		},
		{
			Expected: "POST",
			Actual:   func(endpoint *flux.Endpoint) interface{} { return endpoint.HttpMethod },
		},
		{
			Expected: 0,
			Actual:   func(endpoint *flux.Endpoint) interface{} { return len(endpoint.Permissions) },
		},
		{
			Expected: "",
			Actual:   func(endpoint *flux.Endpoint) interface{} { return endpoint.Attr("roles").ToString() },
		},
		{
			Expected: true,
			Actual:   func(endpoint *flux.Endpoint) interface{} { return endpoint.Attr("Authorize").ToBool() },
		},
		{
			Expected: "com.foo.bar.act.IActivityService:list",
			Actual:   func(endpoint *flux.Endpoint) interface{} { return endpoint.ServiceId },
		},
	})
}

func TestParseEndpointModelV1_0(t *testing.T) {
	text := `{
    "application": "myapp",
    "authorize": true,
    "endpointId": "myapp:/v1/api/seller.adm.activity.list:1.0",
    "extensions": {
        "protoType": "GatewayAtop",
		"biz": [
            "MALL"
        ],
		"role": [
            "APP_USER"
        ]
    },
    "httpMethod": "POST",
    "httpPattern": "/v1/api/seller.adm.activity.list",
    "permissions": [],
    "service": {
        "interface": "com.foo.bar.act.IActivityService",
        "method": "list",
        "rpcGroup": "",
        "rpcProto": "DUBBO",
        "rpcVersion": "",
        "serviceId": "com.foo.bar.act.IActivityService:list"
    },
    "serviceId": "com.foo.bar.act.IActivityService:list",
    "version": "1.0"
}`
	serializer := flux.NewJsonSerializer()
	ext.RegisterSerializer(ext.TypeNameSerializerJson, serializer)
	evt, err := NewEndpointEvent([]byte(text), remoting.EventTypeNodeAdd)
	if nil != err {
		t.Fatal("parse endpoint event failed, err :", err)
	}
	//flux.Assert
	AssertEndpointModel(t, &evt.Endpoint, []AssertCase{
		{
			Expected: true,
			Actual:   func(endpoint *flux.Endpoint) interface{} { return endpoint.IsValid() },
		},
		{
			Expected: "MALL",
			Actual:   func(endpoint *flux.Endpoint) interface{} { return endpoint.Attr("biz").ToString() },
		},
		{
			Expected: []string{"APP_USER"},
			Actual:   func(endpoint *flux.Endpoint) interface{} { return endpoint.Attr("role").ToStringSlice() },
		},
	})
}

func AssertEndpointModel(t *testing.T, endpoint *flux.Endpoint, cases []AssertCase) {
	tAssert := assert2.New(t)
	for _, c := range cases {
		tAssert.Equal(c.Expected, c.Actual(endpoint), c.Message)
	}
}

type AssertCase struct {
	Expected interface{}
	Actual   func(endpoint *flux.Endpoint) interface{}
	Message  string
}
