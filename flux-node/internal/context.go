package internal

type ContextKey struct {
	key string
}

var (
	ContextKeyRequestId     = ContextKey{key: "__internal.context.request.id"}
	ContextKeyRouteEndpoint = ContextKey{key: "__internal.context.route.endpoint"}
)
