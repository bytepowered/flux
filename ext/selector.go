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
	hostedSelectors    = make(map[string][]flux.Selector, 16)
	hostedSelectorLock sync.RWMutex
)

func StoreSelector(s flux.Selector) {
	pkg.RequireNotNil(s, "Selector is nil")
	StoreHostedSelector(anyHost, s)
}

func StoreHostedSelector(host string, s flux.Selector) {
	host = pkg.RequireNotEmpty(host, "host is empty")
	pkg.RequireNotNil(s, "Selector is nil")
	hostedSelectorLock.Lock()
	defer hostedSelectorLock.Unlock()
	if l, ok := hostedSelectors[host]; ok {
		hostedSelectors[host] = append(l, s)
	} else {
		hostedSelectors[host] = []flux.Selector{s}
	}
}

func FindSelectors(host string) []flux.Selector {
	host = pkg.RequireNotEmpty(host, "host is empty")
	hostedSelectorLock.RLock()
	defer hostedSelectorLock.RUnlock()
	if hosted, ok := hostedSelectors[host]; ok {
		return _newSelectors(hosted)
	} else if anyHost != host {
		if any, ok := hostedSelectors[anyHost]; ok {
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
