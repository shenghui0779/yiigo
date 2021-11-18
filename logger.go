package yiigo

import (
	"os"
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

// LoggerOptions options configure a Logger.
type LoggerOptions struct {
	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool

	// Stderr specifies the stderr for logger
	Stderr bool

	// ZapOptions specifies the zap options stderr for logger
	ZapOptions []zap.Option
}

// newLogger returns a new logger.
func newLogger(path string, opt *LoggerOptions) *zap.Logger {
	if len(path) == 0 {
		return debugLogger(opt.ZapOptions...)
	}

	c := zap.NewProductionEncoderConfig()

	c.TimeKey = "time"
	c.EncodeTime = MyTimeEncoder
	c.EncodeCaller = zapcore.FullCallerEncoder

	ws := make([]zapcore.WriteSyncer, 0, 2)

	ws = append(ws, zapcore.AddSync(&lumberjack.Logger{
		Filename:   path,
		MaxSize:    opt.MaxSize,
		MaxBackups: opt.MaxBackups,
		MaxAge:     opt.MaxAge,
		Compress:   opt.Compress,
		LocalTime:  true,
	}))

	if opt.Stderr {
		ws = append(ws, zapcore.Lock(os.Stderr))
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(c), zapcore.NewMultiWriteSyncer(ws...), zap.DebugLevel)

	return zap.New(core, opt.ZapOptions...)
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

func initLogger(name, path string, opt *LoggerOptions) {
	if opt == nil {
		opt = new(LoggerOptions)
	}

	l := newLogger(path, opt)

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
