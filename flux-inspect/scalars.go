package fluxinspect

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/graphql-go/graphql"
)

var (
	// Map
	MapScalarType = graphql.NewScalar(graphql.ScalarConfig{
		Name:        "JSONObject",
		Description: "由服务端单向序列化的JSONObject对象.",
		Serialize: func(value interface{}) interface{} {
			return value
		},
	})
	// Endpoint
	ServiceScalarType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Service",
		Description: "网关后端Service元数据",
		Fields: graphql.Fields{
			"aliasId": newServiceStringField("Service别名", func(srv *flux.Service) string {
				return srv.AliasId
			}),
			"serviceId": newServiceStringField("Service的标识ID", func(srv *flux.Service) string {
				return srv.ServiceID()
			}),
			"scheme": newServiceStringField("Service侧URL的Scheme", func(srv *flux.Service) string {
				return srv.Scheme
			}),
			"url": newServiceStringField("Service侧的Url", func(srv *flux.Service) string {
				return srv.Url
			}),
			"interface": newServiceStringField("Service侧的URL/Interface", func(srv *flux.Service) string {
				return srv.Interface
			}),
			"method": newServiceStringField("Service侧的方法", func(srv *flux.Service) string {
				return srv.Method
			}),
			"attributes": newAttributesField("服务属性列表", func(src interface{}) []flux.Attribute {
				srv, _ := src.(*flux.Service)
				return srv.Attributes
			}),
			"arguments": newArgumentField("接口参数列表"),
		},
	})
	// Endpoint
	EndpointScalarType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Endpoint",
		Description: "已注册到网关的Endpoint元数据",
		Fields: graphql.Fields{
			"application": newEndpointStringField("所属应用名", func(ep *flux.Endpoint) string {
				return ep.Application
			}),
			"version": newEndpointStringField("端点版本号", func(ep *flux.Endpoint) string {
				return ep.Version
			}),
			"httpPattern": newEndpointStringField("映射Http侧的UriPattern", func(ep *flux.Endpoint) string {
				return ep.HttpPattern
			}),
			"httpMethod": newEndpointStringField("映射Http侧的UriPattern", func(ep *flux.Endpoint) string {
				return ep.HttpMethod
			}),
			"attributes": newAttributesField("端点属性列表", func(src interface{}) []flux.Attribute {
				srv, _ := src.(*flux.Endpoint)
				return srv.Attributes
			}),
			"service": &graphql.Field{
				Type:        ServiceScalarType,
				Description: "上游/后端服务",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ep, _ := p.Source.(*flux.Endpoint)
					return &ep.Service, nil
				},
			},
			"permissions": &graphql.Field{
				Type:        MapScalarType,
				Description: "多组权限验证服务ID列表",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ep, _ := p.Source.(*flux.Endpoint)
					if ep.Permissions == nil {
						return []string{}, nil
					}
					return ep.Permissions, nil
				},
			},
		},
	})
)

func newAttributesField(desc string, af func(interface{}) []flux.Attribute) *graphql.Field {
	return &graphql.Field{
		Type:        MapScalarType,
		Description: desc,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			attrs := af(p.Source)
			if attrs == nil {
				attrs = []flux.Attribute{}
			}
			return attrs, nil
		},
	}
}

func newArgumentField(desc string) *graphql.Field {
	return &graphql.Field{
		Type:        MapScalarType,
		Description: desc,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			var args []flux.Argument
			if srv, ok := p.Source.(*flux.Service); ok {
				args = srv.Arguments
			}
			if args == nil {
				args = []flux.Argument{}
			}
			return args, nil
		},
	}
}

func newServiceStringField(desc string, f func(endpoint *flux.Service) string) *graphql.Field {
	return &graphql.Field{
		Type:        graphql.NewNonNull(graphql.String),
		Description: desc,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			if srv, ok := p.Source.(*flux.Service); ok {
				return f(srv), nil
			}
			return "", nil
		},
	}
}

func newEndpointStringField(desc string, f func(endpoint *flux.Endpoint) string) *graphql.Field {
	return &graphql.Field{
		Type:        graphql.NewNonNull(graphql.String),
		Description: desc,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			if ep, ok := p.Source.(*flux.Endpoint); ok {
				return f(ep), nil
			}
			return "", nil
		},
	}
}
