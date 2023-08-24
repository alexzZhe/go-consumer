package mq

import (
	l "example.com/demo/pkg/log"
	"github.com/ThreeDotsLabs/watermill"
	"go.uber.org/zap"
)

type mqlog struct {
	logger *zap.SugaredLogger
}

func newMQLogger() *mqlog {
	return &mqlog{
		logger: l.SugaredLogger(),
	}
}

func (l *mqlog) Error(msg string, err error, fields watermill.LogFields) {
	l.logger.Errorf(msg+" %v %v", fields, err)
}
func (l *mqlog) Info(msg string, fields watermill.LogFields) {
	l.logger.Infof(msg+" %v", fields)
}
func (l *mqlog) Debug(msg string, fields watermill.LogFields) {
	l.logger.Debugf(msg+" %v", fields)
}
func (l *mqlog) Trace(msg string, fields watermill.LogFields) {
	l.logger.Debugf(msg+" %v", fields)
}

func (l *mqlog) With(fields watermill.LogFields) watermill.LoggerAdapter { return l }
