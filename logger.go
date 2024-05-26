package yiigo

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogConfig 日志初始化配置
type LogConfig struct {
	// Filename 日志名称
	Filename string
	// Level 日志级别
	Level zapcore.Level
	// MaxSize 当前文件多大时轮替；默认：100MB
	MaxSize int
	// MaxAge 轮替的旧文件最大保留时长；默认：不限
	MaxAge int
	// MaxBackups 轮替的旧文件最大保留数量；默认：不限
	MaxBackups int
	// Compress 轮替的旧文件是否压缩；默认：不压缩
	Compress bool
	// Stderr 是否输出到控制台
	Stderr bool
	// Options Zap日志选项
	Options []zap.Option
}

func DebugLogger(options ...zap.Option) *zap.Logger {
	cfg := zap.NewDevelopmentConfig()

	cfg.DisableCaller = true
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = MyTimeEncoder
	cfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder

	logger, _ := cfg.Build(options...)
	return logger
}

func NewLogger(cfg *LogConfig) *zap.Logger {
	if len(cfg.Filename) == 0 {
		return DebugLogger(cfg.Options...)
	}

	ec := zap.NewProductionEncoderConfig()
	ec.TimeKey = "time"
	ec.EncodeTime = MyTimeEncoder
	ec.EncodeCaller = zapcore.FullCallerEncoder

	ws := []zapcore.WriteSyncer{
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			LocalTime:  true,
			Compress:   cfg.Compress,
		}),
	}
	if cfg.Stderr {
		ws = append(ws, zapcore.Lock(os.Stderr))
	}
	return zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(ec), zapcore.NewMultiWriteSyncer(ws...), cfg.Level), cfg.Options...)
}

// MyTimeEncoder 自定义时间格式化
func MyTimeEncoder(t time.Time, e zapcore.PrimitiveArrayEncoder) {
	e.AppendString(t.In(time.Local).Format(time.DateTime))
}
