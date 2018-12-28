package yiigo

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type logConf struct {
	Path       string `toml:"path"`
	MaxSize    int    `toml:"maxSize"`
	MaxBackups int    `toml:"maxBackups"`
	MaxAge     int    `toml:"maxAge"`
	Compress   bool   `toml:"compress"`
}

// Logger yiigo logger
var Logger *zap.Logger

// initLogger init logger
func initLogger() {
	if Env.Bool("app.debug", false) {
		cfg := zap.NewDevelopmentConfig()

		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = MyTimeEncoder

		Logger, _ = cfg.Build()
	} else {
		conf := &logConf{
			Path:    "app.log",
			MaxSize: 500,
		}

		Env.Unmarshal("log", conf)

		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   conf.Path,
			MaxSize:    conf.MaxSize, // MB
			MaxBackups: conf.MaxBackups,
			MaxAge:     conf.MaxAge, // days
			Compress:   conf.Compress,
		})

		cfg := zap.NewProductionEncoderConfig()

		cfg.TimeKey = "time"
		cfg.EncodeTime = MyTimeEncoder
		cfg.EncodeCaller = zapcore.FullCallerEncoder

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg),
			w,
			zap.DebugLevel,
		)

		Logger = zap.New(core, zap.AddCaller())
	}
}

// MyTimeEncoder zap time encoder.
func MyTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}
