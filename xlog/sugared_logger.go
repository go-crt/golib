package xlog

import (
	"github.com/gin-gonic/gin"
	"github.com/go-crt/golib/env"
	"go.uber.org/zap"
)

// GetLogger 获得一个新的logger 会把日志打印到 name.log 中，不建议业务使用
// deprecated
func GetLogger() (s *zap.SugaredLogger) {
	if SugaredLogger == nil {
		SugaredLogger = newLogger().WithOptions(zap.AddCallerSkip(1)).Sugar()
	}
	return SugaredLogger
}

// 通用字段封装
func sugaredLogger(ctx *gin.Context) *zap.SugaredLogger {
	if ctx == nil {
		return SugaredLogger
	}

	return SugaredLogger.With(
		zap.String("logId", GetLogID(ctx)),
		zap.String("requestId", GetRequestID(ctx)),
		zap.String("module", env.GetAppName()),
		zap.String("localIp", env.LocalIP),
	)
}

// 提供给业务使用的server log 日志打印方法
func Debug(ctx *gin.Context, args ...interface{}) {
	if NoLog(ctx) {
		return
	}
	sugaredLogger(ctx).Debug(args...)
}

func Debugf(ctx *gin.Context, format string, args ...interface{}) {
	if NoLog(ctx) {
		return
	}
	sugaredLogger(ctx).Debugf(format, args...)
}

func Info(ctx *gin.Context, args ...interface{}) {
	if NoLog(ctx) {
		return
	}
	sugaredLogger(ctx).Info(args...)
}

func Infof(ctx *gin.Context, format string, args ...interface{}) {
	if NoLog(ctx) {
		return
	}
	sugaredLogger(ctx).Infof(format, args...)
}

func Warn(ctx *gin.Context, args ...interface{}) {
	if NoLog(ctx) {
		return
	}
	sugaredLogger(ctx).Warn(args...)
}

func Warnf(ctx *gin.Context, format string, args ...interface{}) {
	if NoLog(ctx) {
		return
	}
	sugaredLogger(ctx).Warnf(format, args...)
}

func Error(ctx *gin.Context, args ...interface{}) {
	if NoLog(ctx) {
		return
	}
	sugaredLogger(ctx).Error(args...)
}

func Errorf(ctx *gin.Context, format string, args ...interface{}) {
	if NoLog(ctx) {
		return
	}
	sugaredLogger(ctx).Errorf(format, args...)
}

func Panic(ctx *gin.Context, args ...interface{}) {
	if NoLog(ctx) {
		return
	}
	sugaredLogger(ctx).Panic(args...)
}

func Panicf(ctx *gin.Context, format string, args ...interface{}) {
	if NoLog(ctx) {
		return
	}
	sugaredLogger(ctx).Panicf(format, args...)
}

func Fatal(ctx *gin.Context, args ...interface{}) {
	if NoLog(ctx) {
		return
	}
	sugaredLogger(ctx).Fatal(args...)
}

func Fatalf(ctx *gin.Context, format string, args ...interface{}) {
	if NoLog(ctx) {
		return
	}
	sugaredLogger(ctx).Fatalf(format, args...)
}
