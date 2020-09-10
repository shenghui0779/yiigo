package yiigo

import (
	"sync"
	"time"

	"github.com/pelletier/go-toml"
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
	tree, ok := env.get("log").(*toml.Tree)

	if !ok {
		return
	}

	keys := tree.Keys()

	if len(keys) == 0 {
		return
	}

	for _, v := range keys {
		node, ok := tree.Get(v).(*toml.Tree)

		if !ok {
			continue
		}

		cfg := &logConfig{
			Path:       "app.log",
			MaxSize:    500,
			MaxBackups: 0,
			MaxAge:     0,
			Compress:   true,
		}

		node.Unmarshal(cfg)

		l := newLogger(cfg, debug)

		if v == AsDefault {
			logger = l
		}

		logMap.Store(v, l)
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
