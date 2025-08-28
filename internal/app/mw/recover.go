package mw

import (
	"errors"
	"runtime/debug"

	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/metric"
	"go.uber.org/zap"
)

func handleRecover(method string, recoverVal any) {
	metric.ServerRequestPanics.Inc()
	var err error
	switch x := recoverVal.(type) {
	case string:
		err = errors.New(x)
	case error:
		err = x
	default:
		err = errors.New("unknown panic")
	}
	logger.Error("recovered after panic",
		zap.String("method", method),
		zap.String("stack_trace", string(debug.Stack())),
		zap.Error(err),
	)
}
