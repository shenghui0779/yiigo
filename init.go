package yiigo

func init() {
	// init default logger
	logger = newLogger(&logConfig{
		Path:       "app.log",
		MaxSize:    500,
		MaxBackups: 0,
		MaxAge:     0,
		Compress:   true,
	}, false)

	// load config file: yiigo.toml
	loadConfigFile()

	debug := Env("app.debug").Bool(true)

	// init logger
	initLogger(debug)
	// init db
	initDB(debug)
	// init mongodb
	initMongoDB()
	// init redis
	initRedis()
	// init mailer
	initMailer()
	// init apollo
	initApollo()
}
