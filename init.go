package yiigo

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

// HTTPOption configures how we set up the yiigo init.
type InitOption func(s *initSettings)

// WithEnvDir specifies the dir to load env.
func WithEnvDir(path string) InitOption {
	return func(s *initSettings) {
		s.envDir = path
	}
}

// WithEnvWatcher specifies watching env change.
func WithEnvWatcher() InitOption {
	return func(s *initSettings) {
		s.envWatcher = true
	}
}

// Init yiigo init
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
