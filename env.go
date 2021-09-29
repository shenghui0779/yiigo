package yiigo

import (
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// EnvOnchangeFunc handles the change of env file
type EnvOnchangeFunc func(event fsnotify.Event)

type environment struct {
	path      string
	watcher   bool
	onchanges []EnvOnchangeFunc
}

func (e *environment) initWatcher() {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		logger.Error("[yiigo] env watcher error", zap.Error(err))

		return
	}

	defer watcher.Close()

	envDir, _ := filepath.Split(e.path)
	realEnvFile, _ := filepath.EvalSymlinks(e.path)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer func() {
			wg.Done()

			if r := recover(); r != nil {
				logger.Error("[yiigo] env watcher panic", zap.Any("error", r), zap.String("env_file", e.path), zap.ByteString("stack", debug.Stack()))
			}
		}()

		writeOrCreateMask := fsnotify.Write | fsnotify.Create

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok { // 'Events' channel is closed
					return
				}

				eventFile := filepath.Clean(event.Name)
				currentEnvFile, _ := filepath.EvalSymlinks(e.path)

				// the env file was modified or created || the real path to the env file changed (eg: k8s ConfigMap replacement)
				if (eventFile == e.path && event.Op&writeOrCreateMask != 0) || (currentEnvFile != "" && currentEnvFile != realEnvFile) {
					realEnvFile = currentEnvFile

					if err := godotenv.Overload(e.path); err != nil {
						logger.Error("[yiigo] env reload error", zap.Error(err), zap.String("env_file", e.path))
					}

					for _, f := range e.onchanges {
						f(event)
					}
				} else if eventFile == e.path && event.Op&fsnotify.Remove&fsnotify.Remove != 0 {
					logger.Warn("[yiigo] env file removed", zap.String("envfile", e.path))
				}
			case err, ok := <-watcher.Errors:
				if ok { // 'Errors' channel is not closed
					logger.Error("[yiigo] env watcher error", zap.Error(err), zap.String("env_file", e.path))
				}

				return
			}
		}
	}()

	watcher.Add(envDir)

	wg.Wait()
}

// EnvOption configures how we set up the env file.
type EnvOption func(e *environment)

// WithEnvFile specifies the env file.
func WithEnvFile(filename string) EnvOption {
	return func(e *environment) {
		if len(strings.TrimSpace(filename)) == 0 {
			return
		}

		e.path = filepath.Clean(filename)
	}
}

// WithEnvWatcher watches the change of env file.
func WithEnvWatcher(fn ...EnvOnchangeFunc) EnvOption {
	return func(e *environment) {
		e.watcher = true
		e.onchanges = fn
	}
}

// LoadEnv will read your env file(s) and load them into ENV for this process.
// It will default to loading .env in the current path if not specifies the filename.
func LoadEnv(options ...EnvOption) error {
	env := &environment{path: ".env"}

	for _, f := range options {
		f(env)
	}

	if err := godotenv.Overload(env.path); err != nil {
		return err
	}

	if env.watcher {
		go env.initWatcher()
	}

	return nil
}
