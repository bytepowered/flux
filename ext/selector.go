package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
	"sync"
)

const (
	anyHost = "*"
)

var (
	_hostedSelectors    = make(map[string][]flux.Selector, 16)
	_hostedSelectorLock sync.RWMutex
)

func StoreSelector(s flux.Selector) {
	pkg.RequireNotNil(s, "Selector is nil")
	StoreHostedSelector(anyHost, s)
}

func StoreHostedSelector(host string, s flux.Selector) {
	host = pkg.RequireNotEmpty(host, "host is empty")
	pkg.RequireNotNil(s, "Selector is nil")
	_hostedSelectorLock.Lock()
	defer _hostedSelectorLock.Unlock()
	if l, ok := _hostedSelectors[host]; ok {
		_hostedSelectors[host] = append(l, s)
	} else {
		_hostedSelectors[host] = []flux.Selector{s}
	}
}

func FindSelectors(host string) []flux.Selector {
	host = pkg.RequireNotEmpty(host, "host is empty")
	_hostedSelectorLock.RLock()
	defer _hostedSelectorLock.RUnlock()
	if hosted, ok := _hostedSelectors[host]; ok {
		return _newSelectors(hosted)
	} else if anyHost != host {
		if any, ok := _hostedSelectors[anyHost]; ok {
			return _newSelectors(any)
		}
	}
	return nil
}

func _newSelectors(src []flux.Selector) []flux.Selector {
	out := make([]flux.Selector, len(src))
	copy(out, src)
	return out
}
