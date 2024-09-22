package cosmoswrappers

import (
	"cosmossdk.io/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	bklog "github.com/berachain/beacon-kit/mod/log"

	"github.com/ava-labs/avalanchego/utils/logging"
)

var _ log.Logger = (*AvaLogWrap)(nil)

type AvaLogWrap struct {
	log logging.Logger
}

func NewAvaLogWrapper(log logging.Logger) *AvaLogWrap {
	return &AvaLogWrap{
		log: log,
	}
}

func (alw *AvaLogWrap) Info(msg string, keyVals ...any) {
	alw.log.Info(msg, toZapFields(keyVals...)...)
}

func (alw *AvaLogWrap) Warn(msg string, keyVals ...any) {
	alw.log.Warn(msg, toZapFields(keyVals...)...)
}

func (alw *AvaLogWrap) Error(msg string, keyVals ...any) {
	alw.log.Error(msg, toZapFields(keyVals...)...)
}

func (alw *AvaLogWrap) Debug(msg string, keyVals ...any) {
	alw.log.Debug(msg, toZapFields(keyVals...)...)
}

func (alw *AvaLogWrap) With(keyVals ...any) log.Logger {
	return alw // TODO: figure out how to implement this
}

func (alw *AvaLogWrap) Impl() any {
	return alw // TODO: figure out how to implement this
}

func (alw *AvaLogWrap) AddKeyColor(key any, color bklog.Color) {
	return // TODO: figure out how to implement this
}

func (alw *AvaLogWrap) AddKeyValColor(key any, val any, color bklog.Color) {
	return // TODO: figure out how to implement this
}

func toZapFields(keyVals ...any) []zapcore.Field {
	fields := make([]zapcore.Field, 0, len(keyVals))
	for _, v := range keyVals {
		fields = append(fields, zap.Any("", v))
	}
	return fields
}
