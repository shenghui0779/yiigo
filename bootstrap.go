package yiigo

// Bootstrap start components
func Bootstrap(mysql bool, mongo bool, redis bool) error {
	loadEnv("env.ini")
	initLogger()

	if mysql {
		if err := initMySQL(); err != nil {
			return err
		}
	}

	if mongo {
		if err := initMongo(); err != nil {
			return err
		}
	}

	if redis {
		if err := initRedis(); err != nil {
			return err
		}
	}

	return nil
}
