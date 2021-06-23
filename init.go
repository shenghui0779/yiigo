package yiigo

import (
	"path/filepath"
	"runtime/debug"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// Debug specifies the debug mode
var Debug bool

type initSettings struct {
	envDir       string
	envWatcher   bool
	envOnChange  func(event fsnotify.Event)
	nsqInit      bool
	nsqConsumers []NSQConsumer
}

// InitOption configures how we set up the yiigo initialization.
type InitOption func(s *initSettings)

// WithEnvDir specifies the dir to load env.
func WithEnvDir(dir string) InitOption {
	return func(s *initSettings) {
		s.envDir = filepath.Clean(dir)
	}
}

// WithEnvWatcher specifies watching env change.
func WithEnvWatcher(onchanges ...func(event fsnotify.Event)) InitOption {
	return func(s *initSettings) {
		s.envWatcher = true

		if len(onchanges) != 0 {
			s.envOnChange = func(event fsnotify.Event) {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("yiigo: env onchange callback panic", zap.Any("error", r), zap.ByteString("stack", debug.Stack()))
					}
				}()

				for _, f := range onchanges {
					f(event)
				}
			}
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
