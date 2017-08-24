package yiigo

type Bootstrap struct {
	Env   string
	Log   string
	Mongo bool
	Redis bool
}

// New 创建bootstrap实例
func New() *Bootstrap {
	return &Bootstrap{
		Env:   "env.ini",
		Mongo: false,
		Redis: false,
	}
}

// SetEnv 设置env配置文件
func (b *Bootstrap) SetEnv(path string) {
	b.Env = path
}

// EnableMongo 启用mongo
func (b *Bootstrap) EnableMongo() {
	b.Mongo = true
}

// EnableRedis 启用redis
func (b *Bootstrap) EnableRedis() {
	b.Redis = true
}

// Bootstrap 启动yiigo组件
func (b *Bootstrap) Bootstrap() error {
	err := loadEnv(b.Env)
	initLogger()

	if err != nil {
		return err
	}

	if err := initMySQL(); err != nil {
		return err
	}

	if b.Mongo {
		if err := initMongo(); err != nil {
			return err
		}
	}

	if b.Redis {
		initRedis()
	}

	return nil
}
