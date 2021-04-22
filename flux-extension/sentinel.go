package fluxext

import (
	"fmt"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/logger"
	"net/http"
)

const TypeIdSentinelFilter = "sentinel_filter"

var _ flux.Filter = new(SentinelFilter)

func NewSentinelFilter() flux.Filter {
	return &SentinelFilter{}
}

// SentinelFilter 服务限流、熔断Filter
type SentinelFilter struct {
}

func (c *SentinelFilter) FilterId() string {
	return TypeIdSentinelFilter
}

func (c *SentinelFilter) Init(config *flux.Configuration) error {
	logger.Info("Sentinel filter initializing")
	// 默认流控配置: Yaml
	cpath := config.GetString("config_path")
	if cpath != "" {
		err := sentinel.InitWithConfigFile(cpath)
		if err != nil {
			return fmt.Errorf("sentinal init failed, error: %w", err)
		}
		logger.Infow("Sentinel Filter init resource", "config", cpath)
	}
	// 默认启用自适应流控，启发因子为 load >= 8
	// https://sentinelguard.io/zh-cn/docs/system-adaptive-protection.html
	_, err := system.LoadRules([]*system.Rule{
		{
			MetricType:   system.Load,
			TriggerCount: 8.0,
			Strategy:     system.BBR,
		},
	})
	return err
}

func (c *SentinelFilter) DoFilter(next flux.FilterInvoker) flux.FilterInvoker {
	return func(ctx *flux.Context) *flux.ServeError {
		err, blocked := sentinel.Entry(ctx.ServiceID(), sentinel.WithTrafficType(base.Inbound))
		if blocked != nil {
			return &flux.ServeError{
				StatusCode: http.StatusServiceUnavailable,
				ErrorCode:  flux.ErrorCodeGatewayCircuited,
				Message:    "CIRCUITED:SERVER_BUSY:DOWNGRADE/SENT",
				CauseError: fmt.Errorf("sentinel ciruited, error: %w", blocked.Error()),
			}
		}
		defer err.Exit()
		return next(ctx)
	}
}
