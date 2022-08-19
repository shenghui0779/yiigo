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

// EnvOnChangeFunc the function that runs each time env change occurs.
type EnvOnChangeFunc func(e fsnotify.Event)

type environment struct {
	path    string
	watcher bool
	eventFn EnvOnChangeFunc
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
func WithEnvWatcher(fn EnvOnChangeFunc) EnvOption {
	return func(e *environment) {
		e.watcher = true
		e.eventFn = fn
	}
}

// LoadEnv will read your env file(s) and load them into ENV for this process.
// It will default to loading .env in the current path if not specifies the filename.
func LoadEnv(options ...EnvOption) {
	env := &environment{path: ".env"}

	for _, f := range options {
		f(env)
	}

	filename, err := filepath.Abs(env.path)

	if err != nil {
		logger.Panic("[yiigo] err load env", zap.Error(err))
	}

	statEnvFile(filename)

	if err := godotenv.Overload(filename); err != nil {
		logger.Panic("[yiigo] err load env", zap.Error(err))
	}

	if env.watcher {
		go watchEnvFile(filename, env.eventFn)
	}
}

func statEnvFile(filename string) {
	if _, err := os.Stat(filename); err == nil {
		return
	}

	if err := os.MkdirAll(path.Dir(filename), 0775); err != nil {
		return
	}

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775)

	if err != nil {
		return
	}

	f.Close()
}

func watchEnvFile(filename string, fn EnvOnChangeFunc) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[yiigo] env watcher panic", zap.Any("error", r), zap.String("env_file", filename), zap.ByteString("stack", debug.Stack()))
		}
	}()

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		logger.Error("[yiigo] err env watcher", zap.Error(err))

		return
	}

	defer watcher.Close()

	done := make(chan error)
	defer close(done)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("[yiigo] env watcher panic", zap.Any("error", r), zap.String("env_file", filename), zap.ByteString("stack", debug.Stack()))
			}
		}()

		realEnvFile, _ := filepath.EvalSymlinks(filename)
		createOrWriteMask := fsnotify.Create | fsnotify.Write

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					done <- errors.New("channel(watcher.Events) is closed")

					return
				}

				eventFile := filepath.Clean(event.Name)

				if eventFile == filename {
					// the env file was created or modified
					if event.Op&createOrWriteMask != 0 {
						if err := godotenv.Overload(filename); err != nil {
							logger.Error("[yiigo] err env reload", zap.Error(err), zap.String("env_file", filename))
						}

						if fn != nil {
							fn(event)
						}
					} else if event.Op&fsnotify.Remove != 0 {
						logger.Warn("[yiigo] env file removed", zap.String("env_file", filename))
					}
				} else {
					currentEnvFile, _ := filepath.EvalSymlinks(filename)

					// the real filename to the env file changed (eg: k8s ConfigMap replacement)
					if len(currentEnvFile) != 0 && currentEnvFile != realEnvFile {
						realEnvFile = currentEnvFile

						if err := godotenv.Overload(filename); err != nil {
							logger.Error("[yiigo] err env reload", zap.Error(err), zap.String("env_file", filename))
						}

						if fn != nil {
							fn(event)
						}
					}
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

	if err = watcher.Add(path.Dir(filename)); err != nil {
		done <- err
	}

	err = <-done

	logger.Error("[yiigo] err env watcher", zap.Error(err), zap.String("env_file", filename))
}
