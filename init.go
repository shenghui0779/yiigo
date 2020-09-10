package yiigo

var debug bool

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

	debug = Env("app.debug").Bool(false)

	// init logger
	initLogger()
	// init db
	initDB()
	// init mongodb
	initMongoDB()
	// init redis
	initRedis()
	// init mailer
	initMailer()
	// init apollo
	initApollo()
}
