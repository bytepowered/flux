package internal

import (
	"github.com/bytepowered/flux"
	"sync"
)

func NewMultiVersionEndpoint() *MultiVersionEndpoint {
	return &MultiVersionEndpoint{
		versions: make(map[string]*flux.Endpoint),
		rwmu:     new(sync.RWMutex),
	}
}

// Multi version Endpoint
type MultiVersionEndpoint struct {
	versions map[string]*flux.Endpoint // 各版本数据
	latest   *flux.Endpoint            // 最新版本
	rwmu     *sync.RWMutex             // 读写锁
}

func (m *MultiVersionEndpoint) Application() string {
	return m.latest.Application
}

func (m *MultiVersionEndpoint) ProtoName() string {
	return m.latest.Protocol
}

func (m *MultiVersionEndpoint) HttpPattern() string {
	return m.latest.HttpPattern
}

func (m *MultiVersionEndpoint) UpstreamUri() string {
	return m.latest.UpstreamUri
}

func (m *MultiVersionEndpoint) Get(version string) (*flux.Endpoint, bool) {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()
	if "" == version {
		return m.latest, true
	}
	if 1 == len(m.versions) {
		for _, v := range m.versions {
			return v, true
		}
		return nil, false
	} else {
		v, ok := m.versions[version]
		return v, ok
	}
}

func (m *MultiVersionEndpoint) Update(version string, endpoint *flux.Endpoint) {
	m.rwmu.Lock()
	m.versions[version] = endpoint
	m.latest = endpoint
	m.rwmu.Unlock()
}

func (m *MultiVersionEndpoint) Delete(version string) {
	m.rwmu.Lock()
	delete(m.versions, version)
	m.rwmu.Unlock()
}

func (m *MultiVersionEndpoint) ToSerializableMap() map[interface{}]interface{} {
	copies := make(map[interface{}]interface{})
	for k, v := range m.versions {
		copies[k] = v
	}
	return copies
}
