package logger

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/go-kratos/kratos/v2/log"
)

type WatermillLogger struct {
	log *log.Helper
}

func NewWatermillLogger(logger log.Logger) *WatermillLogger {
	return &WatermillLogger{log: log.NewHelper(logger)}
}

func (l *WatermillLogger) Error(msg string, err error, fields watermill.LogFields) {
	l.log.Errorf("%s: %v %v", msg, err, fields)
}
func (l *WatermillLogger) Info(msg string, fields watermill.LogFields) {
	l.log.Infof("%s %v", msg, fields)
}
func (l *WatermillLogger) Debug(msg string, fields watermill.LogFields) {
	l.log.Debugf("%s %v", msg, fields)
}
func (l *WatermillLogger) Trace(msg string, fields watermill.LogFields) {
	l.log.Debugf("[TRACE] %s %v", msg, fields)
}
func (l *WatermillLogger) With(fields watermill.LogFields) watermill.LoggerAdapter {
	return l
}
