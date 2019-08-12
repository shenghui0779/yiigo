package yiigo

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Logger default logger
	Logger *zap.Logger
	logMap sync.Map
)

type logOptions struct {
	maxSize    int
	maxAge     int
	maxBackups int
	compress   bool
	debug      bool
}

// LogOption configures how we set up the logger
type LogOption interface {
	apply(options *logOptions)
}

// funcLogOption implements db option
type funcLogOption struct {
	f func(options *logOptions)
}

func (fo *funcLogOption) apply(o *logOptions) {
	fo.f(o)
}

func newFuncLogOption(f func(o *logOptions)) *funcLogOption {
	return &funcLogOption{f: f}
}

// WithLogMaxSize specifies the `MaxSize` to logger.
// MaxSize is the maximum size in megabytes of the log file before it gets
// rotated. It defaults to 100 megabytes.
func WithLogMaxSize(n int) LogOption {
	return newFuncLogOption(func(o *logOptions) {
		o.maxSize = n
	})
}

// WithLogMaxAge specifies the `MaxAge` to logger.
// MaxAge is the maximum number of days to retain old log files based on the
// timestamp encoded in their filename.  Note that a day is defined as 24
// hours and may not exactly correspond to calendar days due to daylight
// savings, leap seconds, etc. The default is not to remove old log files
// based on age.
func WithLogMaxAge(n int) LogOption {
	return newFuncLogOption(func(o *logOptions) {
		o.maxAge = n
	})
}

// WithLogMaxBackups specifies the `MaxBackups` to logger.
// MaxBackups is the maximum number of old log files to retain.  The default
// is to retain all old log files (though MaxAge may still cause them to get
// deleted.)
func WithLogMaxBackups(n int) LogOption {
	return newFuncLogOption(func(o *logOptions) {
		o.maxBackups = n
	})
}

// WithLogCompress specifies the `Compress` to logger.
// Compress determines if the rotated log files should be compressed
// using gzip.
func WithLogCompress(b bool) LogOption {
	return newFuncLogOption(func(o *logOptions) {
		o.compress = b
	})
}

// WithLogDebug specifies the `Debug` mode to logger.
func WithLogDebug(b bool) LogOption {
	return newFuncLogOption(func(o *logOptions) {
		o.debug = b
	})
}

// initLogger init logger, the default `MaxSize` is 500M.
func initLogger(logfile string, options ...LogOption) *zap.Logger {
	o := &logOptions{maxSize: 500}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	if o.debug {
		cfg := zap.NewDevelopmentConfig()

		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = MyTimeEncoder

		l, _ := cfg.Build()

		return l
	}

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logfile,
		MaxSize:    o.maxSize,
		MaxBackups: o.maxBackups,
		MaxAge:     o.maxAge,
		Compress:   o.compress,
	})

	cfg := zap.NewProductionEncoderConfig()

	cfg.TimeKey = "time"
	cfg.EncodeTime = MyTimeEncoder
	cfg.EncodeCaller = zapcore.FullCallerEncoder

	core := zapcore.NewCore(zapcore.NewJSONEncoder(cfg), w, zap.DebugLevel)

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// RegisterLogger register logger
func RegisterLogger(name, file string, options ...LogOption) {
	logger := initLogger(file, options...)

	logMap.Store(name, logger)

	if name == AsDefault {
		Logger = logger
	}
}

// UseLogger returns a logger
func UseLogger(name string) *zap.Logger {
	v, ok := logMap.Load(name)

	if !ok {
		panic(fmt.Errorf("yiigo: logger.%s is not registered", name))
	}

	return v.(*zap.Logger)
}

// MyTimeEncoder zap time encoder.
func MyTimeEncoder(t time.Time, e zapcore.PrimitiveArrayEncoder) {
	e.AppendString(t.Format("2006-01-02 15:04:05"))
}
