package extension

import "github.com/bytepowered/flux"

var (
	_protoNamedExchanges = make(map[string]flux.Exchange)
)

func SetExchange(protoName string, exchange flux.Exchange) {
	_protoNamedExchanges[protoName] = exchange
}

func GetExchange(protoName string) (flux.Exchange, bool) {
	e, ok := _protoNamedExchanges[protoName]
	return e, ok
}

func Exchanges() map[string]flux.Exchange {
	m := make(map[string]flux.Exchange, len(_protoNamedExchanges))
	for p, e := range _protoNamedExchanges {
		m[p] = e
	}
	return m
}
