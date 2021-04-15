package fluxinspect

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux/flux-node"
	fluxpkg "github.com/bytepowered/flux/flux-pkg"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/spf13/cast"
	"net/http"
)

var (
	schema    *graphql.Schema
	schandler flux.WebHandler
)

func init() {
	// Endpoint定义
	endpoints := &graphql.Field{
		Name:        "endpoints",
		Description: "查询已注册到网关的Endpoint列表",
		Type:        graphql.NewList(EndpointScalarType),
		Args: graphql.FieldConfigArgument{
			epQueryKeyApplication: &graphql.ArgumentConfig{
				Description: "通过application过滤特定Endpoint",
				Type:        graphql.String,
			},
			epQueryKeyProtocol: &graphql.ArgumentConfig{
				Description: "通过proto过滤特定Endpoint",
				Type:        graphql.String,
			},
			epQueryKeyPattern: &graphql.ArgumentConfig{
				Description: "通过http-pattern过滤特定Endpoint",
				Type:        graphql.String,
			},
			srvQueryKeyInterface: &graphql.ArgumentConfig{
				Description: "通过service interface过滤特定Endpoint",
				Type:        graphql.String,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return DoQueryEndpoints(func(key string) string {
				return cast.ToString(p.Args[key])
			}), nil
		},
	}
	// Service定义
	services := &graphql.Field{
		Name:        "services",
		Description: "查询已注册到网关的Service列表",
		Type:        graphql.NewList(EndpointScalarType),
		Args: graphql.FieldConfigArgument{
			srvQueryKeyServiceId: &graphql.ArgumentConfig{
				Description: "通过serviceId过滤特定Service",
				Type:        graphql.String,
			},
			srvQueryKeyInterface: &graphql.ArgumentConfig{
				Description: "通过service interface过滤特定Service",
				Type:        graphql.String,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return DoQueryEndpoints(func(key string) string {
				return cast.ToString(p.Args[key])
			}), nil
		},
	}
	sc, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{Name: "RootQuery",
			Fields: graphql.Fields{
				"endpoints": endpoints,
				"services":  services,
			}}),
	})
	fluxpkg.AssertL(err == nil, func() string {
		return fmt.Sprintf("<graphql> scheme init failed: %s", err)
	})
	schema = &sc
}

func NewGraphQLHandlerWith(rootf func(ctx context.Context, r *http.Request) map[string]interface{}) flux.WebHandler {
	if schandler == nil {
		schandler = flux.WrapHttpHandler(handler.New(&handler.Config{
			Schema:       schema,
			RootObjectFn: rootf,
		}))
	}
	return schandler
}

func NewGraphQLHandler() flux.WebHandler {
	return NewGraphQLHandlerWith(nil)
}
