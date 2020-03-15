package extension

import (
	"github.com/bytepowered/flux"
	"sync"
)

const (
	anyHost = "*"
)

var (
	_hostedSelectors map[string][]flux.Selector
	_rwLock          sync.RWMutex
)

func AddSelector(s flux.Selector) {
	AddHostedSelector(anyHost, s)
}

func AddHostedSelector(host string, s flux.Selector) {
	defer _rwLock.Unlock()
	_rwLock.Lock()
	if l, ok := _hostedSelectors[host]; ok {
		_hostedSelectors[host] = append(l, s)
	} else {
		_hostedSelectors[host] = []flux.Selector{s}
	}
}

func FindSelectors(host string) []flux.Selector {
	defer _rwLock.RUnlock()
	_rwLock.RLock()
	if l, ok := _hostedSelectors[host]; ok {
		return _selectors(l)
	} else if anyHost != host {
		if l, ok := _hostedSelectors[anyHost]; ok {
			return _selectors(l)
		}
	}
	return nil
}

func _selectors(l []flux.Selector) []flux.Selector {
	c := make([]flux.Selector, len(l))
	copy(c, l)
	return c
}
