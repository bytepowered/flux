package ext

import "github.com/bytepowered/flux"

var (
	_protoNamedExchanges        = make(map[string]flux.Exchange)
	_protoNamedExchangeDecoders = make(map[string]flux.ExchangeDecoder)
)

func SetExchange(protoName string, exchange flux.Exchange) {
	_protoNamedExchanges[protoName] = exchange
}

func SetExchangeDecoder(protoName string, decoder flux.ExchangeDecoder) {
	_protoNamedExchangeDecoders[protoName] = decoder
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
