package filter

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/lakego"
	"golang.org/x/time/rate"
	"net/http"
	"time"
)

const (
	keyConfigLimitLookupId = "limit-lookup"
	keyConfigLimitRateKey  = "limit-rate"
	keyConfigLimitSizeKey  = "limit-size"
)

const (
	TypeIdRateLimitFilter = "RateLimitFilter"
)

func RateLimitFilterFactory() interface{} {
	return NewRateLimitFilter()
}

func NewRateLimitFilter() flux.Filter {
	return new(RateLimitFilter)
}

type RateLimitConfig struct {
	lookupId  string
	limitRate time.Duration
	limitSize int
}

type RateLimitFilter struct {
	config   *RateLimitConfig
	limiters lakego.Cache
}

func (r *RateLimitFilter) Init(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		keyConfigCacheExpiration: defValueCacheExpiration,
		keyConfigLimitRateKey:    time.Minute,
		keyConfigLimitLookupId:   flux.XJwtSubject,
		keyConfigLimitSizeKey:    1000,
	})
	logger.Info("RateLimit filter initializing")
	r.config = &RateLimitConfig{
		lookupId:  config.GetString(keyConfigLimitLookupId),
		limitRate: config.GetDuration(keyConfigLimitRateKey),
		limitSize: config.GetInt(keyConfigLimitSizeKey),
	}
	// RateLimit缓存大小
	r.limiters = lakego.NewSimple()
	return nil
}

func (r *RateLimitFilter) Invoke(next flux.FilterInvoker) flux.FilterInvoker {
	return func(ctx flux.Context) *flux.InvokeError {
		id := flux.LookupValue(r.config.lookupId, ctx)
		limit, _ := r.limiters.GetOrLoad(id, func(_ lakego.Key) (lakego.Value, error) {
			return rate.NewLimiter(rate.Every(r.config.limitRate), r.config.limitSize), nil
		})
		if limit.(*rate.Limiter).Allow() {
			return next(ctx)
		} else {
			return &flux.InvokeError{
				StatusCode: http.StatusTooManyRequests,
				ErrorCode:  flux.ErrorCodeRequestInvalid,
				Message:    "RATE:OVER_LIMIT",
			}
		}
	}
}

func (*RateLimitFilter) TypeId() string {
	return TypeIdRateLimitFilter
}
