package yiigo

import "path/filepath"

var debug bool

func init() {
	// init default logger
	logger = newLogger(&logConfig{
		Path:       "logs/app.log",
		MaxSize:    500,
		MaxBackups: 0,
		MaxAge:     0,
		Compress:   true,
	})
}

type initSettings struct {
	envDir     string
	envWatcher bool
}

// HTTPOption configures how we set up the yiigo initialization.
type InitOption func(s *initSettings)

// WithEnvDir specifies the dir to load env.
func WithEnvDir(dir string) InitOption {
	return func(s *initSettings) {
		s.envDir = filepath.Clean(dir)
	}
}

// WithEnvWatcher specifies watching env change.
func WithEnvWatcher() InitOption {
	return func(s *initSettings) {
		s.envWatcher = true
	}
}

// Init yiigo initialization
func Init(options ...InitOption) {
	settings := new(initSettings)

	for _, f := range options {
		f(settings)
	}

	initEnv(settings)
	initLogger()
	initDB()
	initMongoDB()
	initRedis()
	initMailer()
}
