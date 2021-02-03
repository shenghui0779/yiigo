package yiigo

import (
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger *zap.Logger
	logMap sync.Map
)

type logConfig struct {
	Path       string `toml:"path"`
	MaxSize    int    `toml:"max_size"`
	MaxBackups int    `toml:"max_backups"`
	MaxAge     int    `toml:"max_age"`
	Compress   bool   `toml:"compress"`
}

// newLogger returns a new logger.
func newLogger(cfg *logConfig, debug bool) *zap.Logger {
	if debug {
		cfg := zap.NewDevelopmentConfig()

		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = MyTimeEncoder

		l, _ := cfg.Build()

		return l
	}

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Path,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	})

	c := zap.NewProductionEncoderConfig()

	c.TimeKey = "time"
	c.EncodeTime = MyTimeEncoder
	c.EncodeCaller = zapcore.FullCallerEncoder

	core := zapcore.NewCore(zapcore.NewJSONEncoder(c), w, zap.DebugLevel)

	return zap.New(core, zap.AddCaller())
}

func initLogger() {
	configs := make(map[string]*logConfig, 0)

	if err := env.Get("log").Unmarshal(&configs); err != nil {
		logger.Panic("yiigo: logger init error", zap.Error(err))
	}

	if len(configs) == 0 {
		return
	}

	for name, cfg := range configs {
		l := newLogger(cfg, debug)

		if name == AsDefault {
			logger = l
		}

		logMap.Store(name, l)
	}
}

// Logger returns a logger
func Logger(name ...string) *zap.Logger {
	if len(name) == 0 {
		return logger
	}

	v, ok := logMap.Load(name[0])

	if !ok {
		return logger
	}

	return v.(*zap.Logger)
}

// MyTimeEncoder zap time encoder.
func MyTimeEncoder(t time.Time, e zapcore.PrimitiveArrayEncoder) {
	e.AppendString(t.Format("2006-01-02 15:04:05"))
}
