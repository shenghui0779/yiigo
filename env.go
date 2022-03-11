package yiigo

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// EnvEventFunc the function that runs each time env change occurs.
type EnvEventFunc func(event fsnotify.Event)

type environment struct {
	path      string
	watcher   bool
	onchanges []EnvEventFunc
}

// EnvOption configures how we set up env file.
type EnvOption func(e *environment)

// WithEnvFile specifies the env file.
func WithEnvFile(filename string) EnvOption {
	return func(e *environment) {
		if v := strings.TrimSpace(filename); len(v) != 0 {
			e.path = filepath.Clean(v)
		}
	}
}

// WithEnvWatcher watching and re-reading env file.
func WithEnvWatcher(fn ...EnvEventFunc) EnvOption {
	return func(e *environment) {
		e.watcher = true
		e.onchanges = append(e.onchanges, fn...)
	}
}

// LoadEnv will read your env file(s) and load them into ENV for this process.
// It will default to loading .env in the current path if not specifies the filename.
func LoadEnv(options ...EnvOption) error {
	env := &environment{path: ".env"}

	for _, f := range options {
		f(env)
	}

	filename, err := filepath.Abs(env.path)

	if err != nil {
		return err
	}

	statEnvFile(filename)

	if err := godotenv.Overload(filename); err != nil {
		return err
	}

	if env.watcher {
		go watchEnvFile(filename, env.onchanges...)
	}

	return nil
}

func statEnvFile(filename string) {
	_, err := os.Stat(filename)

	if err == nil {
		return
	}

	if err = os.MkdirAll(path.Dir(filename), 0775); err != nil {
		return
	}

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0775)

	if err != nil {
		return
	}

	f.Close()
}

func watchEnvFile(path string, onchanges ...EnvEventFunc) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[yiigo] env watcher panic", zap.Any("error", r), zap.String("env_file", path), zap.ByteString("stack", debug.Stack()))
		}
	}()

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		logger.Error("[yiigo] err env watcher", zap.Error(err))

		return
	}

	defer watcher.Close()

	envDir, _ := filepath.Split(path)
	realEnvFile, _ := filepath.EvalSymlinks(path)

	done := make(chan error)
	defer close(done)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("[yiigo] env watcher panic", zap.Any("error", r), zap.String("env_file", path), zap.ByteString("stack", debug.Stack()))
			}
		}()

		writeOrCreateMask := fsnotify.Write | fsnotify.Create

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					done <- errors.New("channel(watcher.Events) is closed")

					return
				}

				eventFile := filepath.Clean(event.Name)
				currentEnvFile, _ := filepath.EvalSymlinks(path)

				// the env file was modified or created || the real path to the env file changed (eg: k8s ConfigMap replacement)
				if (eventFile == path && event.Op&writeOrCreateMask != 0) || (len(currentEnvFile) != 0 && currentEnvFile != realEnvFile) {
					realEnvFile = currentEnvFile

					if err := godotenv.Overload(path); err != nil {
						logger.Error("[yiigo] err env reload", zap.Error(err), zap.String("env_file", path))
					}

					for _, f := range onchanges {
						f(event)
					}
				} else if eventFile == path && event.Op&fsnotify.Remove&fsnotify.Remove != 0 {
					logger.Warn("[yiigo] env file removed", zap.String("env_file", path))
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					err = errors.New("channel(watcher.Errors) is closed")
				}

				done <- err

				return
			}
		}
	}()

	if err = watcher.Add(envDir); err != nil {
		done <- err
	}

	err = <-done

	logger.Error("[yiigo] err env watcher", zap.Error(err), zap.String("env_file", path))
}
