package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config 日志初始化配置
type Config struct {
	// Filename 日志名称
	Filename string
	// Options 日志选项
	Options *Options
}

// Options 日志配置选项
type Options struct {
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
	// ZapOpts Zap日志选项
	ZapOpts []zap.Option
}

func debug(options ...zap.Option) *zap.Logger {
	cfg := zap.NewDevelopmentConfig()

	cfg.DisableCaller = true
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = MyTimeEncoder
	cfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder

	logger, _ := cfg.Build(options...)

	return logger
}

func New(cfg *Config) *zap.Logger {
	if len(cfg.Filename) == 0 {
		return debug()
	}

	ec := zap.NewProductionEncoderConfig()
	ec.TimeKey = "time"
	ec.EncodeTime = MyTimeEncoder
	ec.EncodeCaller = zapcore.FullCallerEncoder

	var zapOpts []zap.Option

	w := &lumberjack.Logger{
		Filename:  cfg.Filename,
		LocalTime: true,
	}
	ws := make([]zapcore.WriteSyncer, 0, 2)
	if cfg.Options != nil {
		zapOpts = cfg.Options.ZapOpts

		w.MaxSize = cfg.Options.MaxSize
		w.MaxAge = cfg.Options.MaxAge
		w.MaxBackups = cfg.Options.MaxBackups
		w.Compress = cfg.Options.Compress

		if cfg.Options.Stderr {
			ws = append(ws, zapcore.Lock(os.Stderr))
		}
	}
	ws = append(ws, zapcore.AddSync(w))

	return zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(ec), zapcore.NewMultiWriteSyncer(ws...), zap.DebugLevel), zapOpts...)
}

// MyTimeEncoder 自定义时间格式化
func MyTimeEncoder(t time.Time, e zapcore.PrimitiveArrayEncoder) {
	e.AppendString(t.In(time.FixedZone("CST", 8*3600)).Format(time.DateTime))
}
