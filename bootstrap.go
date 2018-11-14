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
	// init http client
	initHTTPClient()
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

	errCount := len(ch)

	if errCount > 0 {
		errMsgs := make([]string, 0, errCount)

		for i := 0; i < errCount; i++ {
			err := <-ch
			errMsgs = append(errMsgs, err.Error())
		}

		Logger.Panic(strings.Join(errMsgs, "\n"))
	}
}
