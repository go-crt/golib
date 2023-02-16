package xlog

import (
	"github.com/go-crt/golib/env"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const (
	// trace 日志前缀标识（放在[]zap.Field的第一个位置提高效率）
	TopicType = "_tp"
	// 业务日志名字
	LogNameServer = "server"
	// access 日志文件名字
	LogNameAccess = "access"
	// module 日志文件名字
	LogNameModule = "module"
)

// RegisterXJSONEncoder registers a special jsonEncoder under "x-json" name.
func RegisterXJSONEncoder() error {
	return zap.RegisterEncoder("x-json", func(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
		return NewXJSONEncoder(cfg), nil
	})
}

type jsonHexEncoder struct {
	zapcore.Encoder
}

func NewXJSONEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	jsonEncoder := zapcore.NewJSONEncoder(cfg)
	return &jsonHexEncoder{
		Encoder: jsonEncoder,
	}
}
func (enc *jsonHexEncoder) Clone() zapcore.Encoder {
	encoderClone := enc.Encoder.Clone()
	return &jsonHexEncoder{Encoder: encoderClone}
}
func (enc *jsonHexEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// 增加 trace 日志前缀，ex： tt= tp=module.log
	fName := "server"
	if len(fields) > 1 && fields[0].Key == TopicType {
		fName = fields[0].String // 确保一定是string类型的
		fields = fields[1:]
	}

	buf, err := enc.Encoder.EncodeEntry(ent, fields)
	if !env.IsDockerPlatform() || buf == nil {
		return buf, err
	}

	tp := appendLogFileTail(fName, getLevelType(ent.Level))
	prefix := "tt= tp=" + tp + " "
	n := append([]byte(prefix), buf.Bytes()...)
	buf.Reset()
	_, _ = buf.Write(n)
	return buf, err
}

func getLevelType(lel zapcore.Level) string {
	if lel <= zapcore.InfoLevel {
		return txtLogNormal
	}
	return txtLogWarnFatal
}
