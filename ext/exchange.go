package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	_protoNamedExchanges                = make(map[string]flux.Exchange, 4)
	_protoNamedExchangeResponseDecoders = make(map[string]flux.ExchangeResponseDecoder, 4)
)

func SetExchange(protoName string, exchange flux.Exchange) {
	_protoNamedExchanges[protoName] = pkg.RequireNotNil(exchange, "Exchange is nil").(flux.Exchange)
}

func SetExchangeResponseDecoder(protoName string, decoder flux.ExchangeResponseDecoder) {
	_protoNamedExchangeResponseDecoders[protoName] = pkg.RequireNotNil(decoder, "ExchangeResponseDecoder is nil").(flux.ExchangeResponseDecoder)
}

func GetExchange(protoName string) (flux.Exchange, bool) {
	exchange, ok := _protoNamedExchanges[protoName]
	return exchange, ok
}

func GetExchangeResponseDecoder(protoName string) (flux.ExchangeResponseDecoder, bool) {
	decoder, ok := _protoNamedExchangeResponseDecoders[protoName]
	return decoder, ok
}

func Exchanges() map[string]flux.Exchange {
	m := make(map[string]flux.Exchange, len(_protoNamedExchanges))
	for p, e := range _protoNamedExchanges {
		m[p] = e
	}
	return m
}
