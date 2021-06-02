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

var _ flux.EndpointDiscovery = new(ResourceEndpointDiscovery)

type (
	// ResourceOption 配置函数
	ResourceOption func(discovery *ResourceEndpointDiscovery)
)

type Resources struct {
	Endpoints []flux.Endpoint `yaml:"endpoints"`
	Services  []flux.Service  `yaml:"services"`
}

// NewResourceEndpointDiscovery returns new a resource based discovery service
func NewResourceEndpointDiscovery(id string, opts ...ResourceOption) *ResourceEndpointDiscovery {
	r := &ResourceEndpointDiscovery{
		id:        id,
		resources: make([]Resources, 0, 8),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

type ResourceEndpointDiscovery struct {
	id        string
	resources []Resources
}

func (d *ResourceEndpointDiscovery) Id() string {
	return d.id
}

func (d *ResourceEndpointDiscovery) OnInit(config *flux.Configuration) error {
	// 加载指定路径的配置
	files := config.GetStringSlice("includes")
	logger.Infow("DISCOVERY:RESOURCE:LOAD/resource", "includes", files)
	if err := d.includes(files); nil != err {
		return err
	}
	// 本地指定
	const segEndpoint = "endpoints"
	const segService = "services"
	define := map[string]interface{}{
		segEndpoint: config.GetOrDefault(segEndpoint, make([]interface{}, 0)),
		segService:  config.GetOrDefault(segService, make([]interface{}, 0)),
	}
	if bytes, err := ext.JSONMarshal(define); nil != err {
		return fmt.Errorf("response discovery, redecode config, error: %w", err)
	} else {
		var out Resources
		if err := yaml.Unmarshal(bytes, &out); nil != err {
			return fmt.Errorf("discovery service decode config, err: %w", err)
		} else if len(out.Endpoints) > 0 || len(out.Services) > 0 {
			d.resources = append(d.resources, out)
		}
	}
	return nil
}

func (d *ResourceEndpointDiscovery) WatchEndpoints(ctx context.Context, events chan<- flux.EndpointEvent) error {
	for _, res := range d.resources {
		for _, el := range res.Endpoints {
			if !el.IsValid() {
				logger.Warnw("DISCOVERY:RESOURCE:ENDPOINT/verify", "endpoint", el)
				continue
			}
			dup := el
			if evt, err := ToEndpointEvent(&dup, flux.EventTypeAdded); err == nil {
				events <- evt
			} else {
				logger.Warnw("DISCOVERY:RESOURCE:ENDPOINT/event", "endpoint", el, "error", el)
			}
		}
	}
	return nil
}

func (d *ResourceEndpointDiscovery) WatchServices(ctx context.Context, events chan<- flux.ServiceEvent) error {
	for _, res := range d.resources {
		for _, el := range res.Services {
			if !el.IsValid() {
				logger.Warnw("DISCOVERY:RESOURCE:SERVICE/invalid", "service", el)
				continue
			}
			dup := el
			if evt, err := ToServiceEvent(&dup, flux.EventTypeAdded); err == nil {
				events <- evt
			} else {
				logger.Warnw("DISCOVERY:RESOURCE:SERVICE/event", "service", el, "error", el)
			}
		}
	}
	return nil
}

func (d *ResourceEndpointDiscovery) includes(files []string) error {
	for _, file := range files {
		bytes, err := ioutil.ReadFile(file)
		if nil != err {
			return fmt.Errorf("discovery service read config, path: %s, err: %w", file, err)
		}
		var out Resources
		if err := yaml.Unmarshal(bytes, &out); nil != err {
			return fmt.Errorf("discovery service decode config, path: %s, err: %w", file, err)
		} else {
			d.resources = append(d.resources, out)
		}
	}
	return nil
}