package mw

import (
	"errors"
	"runtime/debug"

	"go.uber.org/zap"

	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/metric"
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
