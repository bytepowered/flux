package echoex

import (
	"github.com/bytepowered/flux"
)

var _echoExchange = new(exchange)

func NewEchoExchange() flux.Exchange {
	return _echoExchange
}

type exchange int

func (e *exchange) Exchange(ctx flux.Context) *flux.InvokeError {
	ep := ctx.Endpoint()
	ret, _ := e.Invoke(&ep, ctx)
	ctx.ResponseWriter().SetStatusCode(flux.StatusOK).SetBody(ret)
	return nil
}

func (e *exchange) Invoke(target *flux.Endpoint, ctx flux.Context) (interface{}, *flux.InvokeError) {
	return map[string]interface{}{
		"endpoint": target,
		"attrs":    ctx.AttrValues(),
		"headers":  ctx.RequestReader().Headers(),
	}, nil
}
