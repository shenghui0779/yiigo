package yiigo

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger = debugLogger()
	logMap = make(map[string]*zap.Logger)
)

// LoggerConfig 日志初始化配置
type LoggerConfig struct {
	// Filename 日志名称
	Filename string `json:"filename"`

	// Options 日志选项
	Options *LoggerOptions `json:"options"`
}

// LoggerOptions 日志配置选项
type LoggerOptions struct {
	// MaxSize 当前文件多大时轮替；默认：100MB
	MaxSize int `json:"max_size"`

	// MaxAge 轮替的旧文件最大保留时长；默认：不限
	MaxAge int `json:"max_age"`

	// MaxBackups 轮替的旧文件最大保留数量；默认：不限
	MaxBackups int `json:"max_backups"`

	// Compress 轮替的旧文件是否压缩；默认：不压缩
	Compress bool `json:"compress"`

	// Stderr 是否输出到控制台
	Stderr bool `json:"stderr"`

	// ZapOpts Zap日志选项
	ZapOpts []zap.Option `json:"zap_opts"`
}

func debugLogger(options ...zap.Option) *zap.Logger {
	cfg := zap.NewDevelopmentConfig()

	cfg.DisableCaller = true
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = MyTimeEncoder
	cfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder

	l, _ := cfg.Build(options...)

	return l
}

func newLogger(cfg *LoggerConfig) *zap.Logger {
	if len(cfg.Filename) == 0 {
		return debugLogger(cfg.Options.ZapOpts...)
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

func initLogger(name string, cfg *LoggerConfig) {
	if cfg.Options == nil {
		cfg.Options = new(LoggerOptions)
	}

	l := newLogger(cfg)
	if name == Default {
		logger = l
	}

	logMap[name] = l
}

// Logger 返回一个日志实例
func Logger(name ...string) *zap.Logger {
	key := Default
	if len(name) != 0 {
		key = name[0]
	}

	l, ok := logMap[key]
	if !ok {
		return logger
	}

	return l
}

// MyTimeEncoder 自定义时间格式化
func MyTimeEncoder(t time.Time, e zapcore.PrimitiveArrayEncoder) {
	e.AppendString(t.In(GMT8).Format(time.DateTime))
}
