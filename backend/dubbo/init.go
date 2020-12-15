package dubbo

import (
	_ "github.com/apache/dubbo-go/cluster/cluster_impl"
	_ "github.com/apache/dubbo-go/cluster/loadbalance"
	_ "github.com/apache/dubbo-go/common/proxy/proxy_factory"
	_ "github.com/apache/dubbo-go/filter/filter_impl"
	_ "github.com/apache/dubbo-go/registry/nacos"
	_ "github.com/apache/dubbo-go/registry/protocol"
	_ "github.com/apache/dubbo-go/registry/zookeeper"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
)

func init() {
	ext.StoreBackendTransport(flux.ProtoDubbo, NewDubboBackendTransport())
	ext.StoreBackendTransportDecodeFunc(flux.ProtoDubbo, NewDubboBackendTransportDecodeFunc())
}
