package yiigo

import (
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger = debugLogger()
	logMap sync.Map
)

type loggerSetting struct {
	maxSize    int
	maxBackups int
	maxAge     int
	compress   bool
}

// LoggerOption configures how we set up the logger.
type LoggerOption func(s *loggerSetting)

// WithLogMaxSize specifies the `MaxSize(Mi)` for logger.
func WithLogMaxSize(n int) LoggerOption {
	return func(s *loggerSetting) {
		s.maxSize = n
	}
}

// WithLogMaxBackups specifies the `MaxBackups` for logger.
func WithLogMaxBackups(n int) LoggerOption {
	return func(s *loggerSetting) {
		s.maxBackups = n
	}
}

// WithLogMaxAge specifies the `MaxAge(days)` for logger.
func WithLogMaxAge(n int) LoggerOption {
	return func(s *loggerSetting) {
		s.maxAge = n
	}
}

// WithLogCompress specifies the `Compress` for logger.
func WithLogCompress() LoggerOption {
	return func(s *loggerSetting) {
		s.compress = true
	}
}

// newLogger returns a new logger.
func newLogger(path string, setting *loggerSetting) *zap.Logger {
	if len(strings.TrimSpace(path)) == 0 {
		return debugLogger()
	}

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path,
		MaxSize:    setting.maxSize,
		MaxBackups: setting.maxBackups,
		MaxAge:     setting.maxAge,
		Compress:   setting.compress,
		LocalTime:  true,
	})

	c := zap.NewProductionEncoderConfig()

	c.TimeKey = "time"
	c.EncodeTime = MyTimeEncoder
	c.EncodeCaller = zapcore.FullCallerEncoder

	core := zapcore.NewCore(zapcore.NewJSONEncoder(c), w, zap.DebugLevel)

	return zap.New(core, zap.AddCaller())
}

func debugLogger() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()

	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = MyTimeEncoder

	l, _ := cfg.Build()

	return l
}

func initLogger(name, path string, options ...LoggerOption) {
	setting := &loggerSetting{
		maxSize: 100,
	}

	for _, f := range options {
		f(setting)
	}

	l := newLogger(path, setting)

	if name == Default {
		logger = l
	}

	logMap.Store(name, l)
}

// Logger returns a logger
func Logger(name ...string) *zap.Logger {
	if len(name) == 0 || name[0] == Default {
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
	e.AppendString(t.Local().Format("2006-01-02 15:04:05"))
}
