package yiigo

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// Logger yiigo logger
var Logger *zap.Logger

// initLogger init logger
func initLogger() {
	if EnvBool("app", "debug", true) {
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   EnvString("log", "path", "app.log"),
			MaxSize:    EnvInt("log.rotate", "maxSize", 500), // megabytes
			MaxBackups: EnvInt("log.rotate", "maxBackups", 0),
			MaxAge:     EnvInt("log.rotate", "maxAge", 0), // days
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
	} else {
		cfg := zap.NewDevelopmentConfig()

		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = MyTimeEncoder

		Logger, _ = cfg.Build()
	}
}

// MyTimeEncoder diy time encoder
func MyTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}
