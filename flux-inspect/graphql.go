package fluxinspect

import (
	"context"
	"github.com/bytepowered/flux/flux-node"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/spf13/cast"
	"net/http"
)

var (
	schema    *graphql.Schema
	schandler flux.WebHandler
)

// InitSchema 初始化GraphQL服务
func InitSchema() error {
	fields := graphql.Fields{
		"endpoints": &graphql.Field{
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
		},
		"services": &graphql.Field{
			Name:        "service",
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
		},
	}
	query := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	sc, err := graphql.NewSchema(graphql.SchemaConfig{Query: graphql.NewObject(query)})
	if err != nil {
		return err
	}
	schema = &sc
	return nil
}

func NewGraphQLHandler() flux.WebHandler {
	if schandler == nil {
		schandler = flux.WrapHttpHandler(handler.New(&handler.Config{
			Schema: schema,
			RootObjectFn: func(ctx context.Context, r *http.Request) map[string]interface{} {
				return map[string]interface{}{
					"webcontext": "webctxref",
				}
			},
		}))
	}
	return schandler
}
