package log

import (
	"github.com/shenghui0779/yiigo"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

var logger = yiigo.DebugLogger()

// Init 初始化日志实例(如有多个实例，在此方法中初始化)
func Init() {
	cfg := &yiigo.LogConfig{
		Filename: viper.GetString("log.path"),
	}

	opts := viper.GetStringMap("log.options")
	if len(opts) != 0 {
		cfg.MaxSize = cast.ToInt(opts["max_size"])
		cfg.MaxAge = cast.ToInt(opts["max_age"])
		cfg.MaxBackups = cast.ToInt(opts["max_backups"])
		cfg.Compress = cast.ToBool(opts["compress"])
		cfg.Stderr = cast.ToBool(opts["stderr"])
	}

	logger = yiigo.NewLogger(cfg)
}
