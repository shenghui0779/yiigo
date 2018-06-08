package yiigo

func init() {
	// load config
	loadEnv("env.toml")
	// init logger
	initLogger()
}

// Bootstrap init and start core components of yiigo.
func Bootstrap(mysql bool, mongo bool, redis bool) error {
	if mysql {
		// init mysql
		if err := initMySQL(); err != nil {
			return err
		}
	}

	if mongo {
		// init mongodb
		if err := initMongo(); err != nil {
			return err
		}
	}

	if redis {
		// init redis
		if err := initRedis(); err != nil {
			return err
		}
	}

	return nil
}
