package internal

type ContextKey struct {
	descriptor string
}

var (
	ContextKeyRequestId     = ContextKey{descriptor: "__internal.context.request.id"}
	ContextKeyRouteEndpoint = ContextKey{descriptor: "__internal.context.route.endpoint"}
)
