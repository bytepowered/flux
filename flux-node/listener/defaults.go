package listener

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/common"
	"github.com/bytepowered/flux/flux-node/logger"
)

// DefaultNotfoundHandler 生成NotFound错误，由ErrorHandler处理
func DefaultNotfoundHandler(_ *flux.WebExchange) error {
	return &flux.ServeError{
		StatusCode: flux.StatusNotFound,
		ErrorCode:  flux.ErrorCodeRequestNotFound,
		Message:    flux.ErrorMessageWebServerRequestNotFound,
	}
}

func DefaultErrorHandler(webex *flux.WebExchange, error error) {
	if nil == error {
		return
	}
	serr, ok := error.(*flux.ServeError)
	if !ok {
		serr = &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    error.Error(),
			CauseError: error,
		}
	}
	if bytes, err := common.SerializeObject(serr); nil != err {
		logger.Trace(webex.RequestId()).Errorw("SERVER:ERROR_HANDLE", "error", err)
	} else if err := webex.Write(serr.StatusCode, flux.MIMEApplicationJSON, bytes); nil != err {
		logger.Trace(webex.RequestId()).Errorw("SERVER:ERROR_HANDLE", "error", err)
	}
}
