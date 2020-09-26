package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	_protoNamedExchanges        = make(map[string]flux.Exchange, 4)
	_protoNamedExchangeDecoders = make(map[string]flux.ExchangeDecoder, 4)
)

func SetExchange(protoName string, exchange flux.Exchange) {
	_protoNamedExchanges[protoName] = pkg.RequireNotNil(exchange, "Exchange is nil").(flux.Exchange)
}

func SetExchangeDecoder(protoName string, decoder flux.ExchangeDecoder) {
	_protoNamedExchangeDecoders[protoName] = pkg.RequireNotNil(decoder, "ExchangeDecoder is nil").(flux.ExchangeDecoder)
}

func GetExchange(protoName string) (flux.Exchange, bool) {
	exchange, ok := _protoNamedExchanges[protoName]
	return exchange, ok
}

func GetExchangeDecoder(protoName string) (flux.ExchangeDecoder, bool) {
	decoder, ok := _protoNamedExchangeDecoders[protoName]
	return decoder, ok
}

func Exchanges() map[string]flux.Exchange {
	m := make(map[string]flux.Exchange, len(_protoNamedExchanges))
	for p, e := range _protoNamedExchanges {
		m[p] = e
	}
	return m
}
