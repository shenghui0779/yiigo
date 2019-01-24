package yiigo

import (
	"strings"
	"sync"
)

func init() {
	// load config
	loadEnv()
	// init logger
	initLogger()
	// init core components
	bootstrap()
}

// bootstrap init core components to yiigo.
func bootstrap() {
	var wg sync.WaitGroup

	ch := make(chan error, 4)

	defer close(ch)

	wg.Add(4)

	// init mysql
	go func(ch chan error) {
		defer wg.Done()

		if err := initMySQL(); err != nil {
			ch <- err

			return
		}
	}(ch)

	// init postgres
	go func(ch chan error) {
		defer wg.Done()

		if err := initPostgres(); err != nil {
			ch <- err

			return
		}
	}(ch)

	// init mongodb
	go func(ch chan error) {
		defer wg.Done()

		if err := initMongo(); err != nil {
			ch <- err

			return
		}
	}(ch)

	// init redis
	go func(ch chan error) {
		defer wg.Done()

		if err := initRedis(); err != nil {
			ch <- err

			return
		}
	}(ch)

	wg.Wait()

	if l := len(ch); l > 0 {
		msgs := make([]string, 0, l)

		for i := 0; i < l; i++ {
			err := <-ch
			msgs = append(msgs, err.Error())
		}

		Logger.Panic(strings.Join(msgs, "\n"))
	}
}
