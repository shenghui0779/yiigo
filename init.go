package yiigo

import (
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

var Debug bool

type initSettings struct {
	envDir       string
	envWatcher   bool
	envOnChange  func(event fsnotify.Event)
	nsqInit      bool
	nsqConsumers []NSQConsumer
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
func WithEnvWatcher(onchange ...func(event fsnotify.Event)) InitOption {
	return func(s *initSettings) {
		s.envWatcher = true

		if len(onchange) != 0 {
			s.envOnChange = onchange[0]
		}
	}
}

// WithNSQ specifies initialize the nsq
func WithNSQ(consumers ...NSQConsumer) InitOption {
	return func(s *initSettings) {
		s.nsqInit = true
		s.nsqConsumers = consumers
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

	if settings.nsqInit {
		initNSQ(settings.nsqConsumers...)
	}
}
