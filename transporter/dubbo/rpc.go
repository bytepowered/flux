package dubbo

import (
	"context"
	"github.com/apache/dubbo-go/protocol"
)

// ResultRPCService uses for generic invoke for service call, returns a raw protocol.Result
type ResultRPCService struct {
	Invoke       func(ctx context.Context, args []interface{}) protocol.Result `dubbo:"$invoke"`
	referenceStr string
}

// NewResultRPCService returns a GenericService instance
func NewResultRPCService(referenceStr string) *ResultRPCService {
	return &ResultRPCService{referenceStr: referenceStr}
}

// Reference gets referenceStr from GenericService
func (u *ResultRPCService) Reference() string {
	return u.referenceStr
}
