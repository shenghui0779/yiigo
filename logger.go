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

	// ZapOptions Zap日志选项
	ZapOptions []zap.Option `json:"zap_options"`
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
		return debugLogger(cfg.Options.ZapOptions...)
	}

	c := zap.NewProductionEncoderConfig()

	c.TimeKey = "time"
	c.EncodeTime = MyTimeEncoder
	c.EncodeCaller = zapcore.FullCallerEncoder

	ws := make([]zapcore.WriteSyncer, 0, 2)

	ws = append(ws, zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.Options.MaxSize,
		MaxAge:     cfg.Options.MaxAge,
		MaxBackups: cfg.Options.MaxBackups,
		Compress:   cfg.Options.Compress,
		LocalTime:  true,
	}))

	if cfg.Options.Stderr {
		ws = append(ws, zapcore.Lock(os.Stderr))
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(c), zapcore.NewMultiWriteSyncer(ws...), zap.DebugLevel)

	return zap.New(core, cfg.Options.ZapOptions...)
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
