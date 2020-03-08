package extension

import "github.com/bytepowered/flux"

var (
	_protoNamedExchanges = make(map[string]flux.Exchange)
)

func SetExchange(protocol string, exchange flux.Exchange) {
	_protoNamedExchanges[protocol] = exchange
}

func GetExchange(protocol string) (flux.Exchange, bool) {
	e, ok := _protoNamedExchanges[protocol]
	return e, ok
}

func Exchanges() map[string]flux.Exchange {
	m := make(map[string]flux.Exchange, len(_protoNamedExchanges))
	for p, e := range _protoNamedExchanges {
		m[p] = e
	}
	return m
}
