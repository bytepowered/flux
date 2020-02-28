package extension

import "github.com/bytepowered/flux"

var (
	_protoExchanges = make(map[string]flux.Exchange)
)

func SetExchange(protocol string, exchange flux.Exchange) {
	_protoExchanges[protocol] = exchange
}

func GetExchange(protocol string) (flux.Exchange, bool) {
	e, ok := _protoExchanges[protocol]
	return e, ok
}

func Exchanges() map[string]flux.Exchange {
	m := make(map[string]flux.Exchange, len(_protoExchanges))
	for p, e := range _protoExchanges {
		m[p] = e
	}
	return m
}
