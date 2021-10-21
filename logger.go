package yiigo

import (
	"os"
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
	stderr     bool
	options    []zap.Option
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

// WithLogStdErr specifies stderr output for logger.
func WithLogStdErr() LoggerOption {
	return func(s *loggerSetting) {
		s.stderr = true
	}
}

func WithZapOptions(options ...zap.Option) LoggerOption {
	return func(s *loggerSetting) {
		s.options = append(s.options, options...)
	}
}

// newLogger returns a new logger.
func newLogger(path string, setting *loggerSetting) *zap.Logger {
	if len(strings.TrimSpace(path)) == 0 {
		return debugLogger(setting.options...)
	}

	c := zap.NewProductionEncoderConfig()

	c.TimeKey = "time"
	c.EncodeTime = MyTimeEncoder
	c.EncodeCaller = zapcore.FullCallerEncoder

	ws := make([]zapcore.WriteSyncer, 0, 2)

	ws = append(ws, zapcore.AddSync(&lumberjack.Logger{
		Filename:   path,
		MaxSize:    setting.maxSize,
		MaxBackups: setting.maxBackups,
		MaxAge:     setting.maxAge,
		Compress:   setting.compress,
		LocalTime:  true,
	}))

	if setting.stderr {
		ws = append(ws, zapcore.Lock(os.Stderr))
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(c), zapcore.NewMultiWriteSyncer(ws...), zap.DebugLevel)

	return zap.New(core, setting.options...)
}

func debugLogger(options ...zap.Option) *zap.Logger {
	cfg := zap.NewDevelopmentConfig()

	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = MyTimeEncoder
	cfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder

	l, _ := cfg.Build(options...)

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
