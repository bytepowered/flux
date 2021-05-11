package discovery

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	ResourceId = "resource"
)

var _ flux.EndpointDiscovery = new(ResourceDiscoveryService)

type (
	// ZookeeperOption 配置函数
	ResourceOption func(discovery *ResourceDiscoveryService)
)

type Resources struct {
	Endpoints []flux.Endpoint `yaml:"endpoints"`
	Services  []flux.Service  `yaml:"services"`
}

// NewResourceServiceWith returns new a resource based discovery service
func NewResourceServiceWith(id string, opts ...ResourceOption) *ResourceDiscoveryService {
	r := &ResourceDiscoveryService{
		id:        id,
		resources: make([]Resources, 0, 8),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

type ResourceDiscoveryService struct {
	id        string
	resources []Resources
}

func (r *ResourceDiscoveryService) Id() string {
	return r.id
}

func (r *ResourceDiscoveryService) Init(config *flux.Configuration) error {
	// 加载指定路径的配置
	files := config.GetStringSlice("includes")
	logger.Infow("Resource discovery, load resources", "includes", files)
	if err := r.includes(files); nil != err {
		return err
	}
	// 本地指定
	define := map[string]interface{}{
		"endpoints": config.GetOrDefault("endpoints", make([]interface{}, 0)),
		"services":  config.GetOrDefault("services", make([]interface{}, 0)),
	}
	if bytes, err := ext.JSONMarshal(define); nil != err {
		return fmt.Errorf("response discovery, redecode config, error: %w", err)
	} else {
		var out Resources
		if err := yaml.Unmarshal(bytes, &out); nil != err {
			return fmt.Errorf("discovery service decode config, err: %w", err)
		} else if len(out.Endpoints) > 0 || len(out.Services) > 0 {
			r.resources = append(r.resources, out)
		}
	}
	return nil
}

func (r *ResourceDiscoveryService) WatchEndpoints(ctx context.Context, events chan<- flux.EndpointEvent) error {
	for _, res := range r.resources {
		for _, ep := range res.Endpoints {
			if ep.IsValid() {
				events <- flux.EndpointEvent{EventType: flux.EventTypeAdded, Endpoint: ep}
			}
		}
	}
	return nil
}

func (r *ResourceDiscoveryService) WatchServices(ctx context.Context, events chan<- flux.ServiceEvent) error {
	for _, res := range r.resources {
		for _, srv := range res.Services {
			if srv.IsValid() {
				events <- flux.ServiceEvent{EventType: flux.EventTypeAdded, Service: srv}
			}
		}
	}
	return nil
}

func (r *ResourceDiscoveryService) includes(files []string) error {
	for _, file := range files {
		bytes, err := ioutil.ReadFile(file)
		if nil != err {
			return fmt.Errorf("discovery service read config, path: %s, err: %w", file, err)
		}
		var out Resources
		if err := yaml.Unmarshal(bytes, &out); nil != err {
			return fmt.Errorf("discovery service decode config, path: %s, err: %w", file, err)
		} else {
			r.resources = append(r.resources, out)
		}
	}
	return nil
}
