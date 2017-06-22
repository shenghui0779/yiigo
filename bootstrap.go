package yiigo

type Bootstrap struct {
	Env   string
	Log   string
	MySQL []string
	Mongo bool
	Redis bool
}

// New 创建bootstrap实例
func New() *Bootstrap {
	return &Bootstrap{
		Env:   "env.ini",
		Log:   "log.xml",
		MySQL: []string{"mysql"},
		Mongo: false,
		Redis: false,
	}
}

// SetEnv 设置env配置文件
func (b *Bootstrap) SetEnv(path string) {
	b.Env = path
}

// SetLog 设置log配置文件
func (b *Bootstrap) SetLog(path string) {
	b.Log = path
}

// SetMySQL 设置mysql配置
func (b *Bootstrap) SetMySQL(sections ...string) {
	b.MySQL = append(b.MySQL, sections...)
}

// EnableMongo 启用mongo
func (b *Bootstrap) EnableMongo() {
	b.Mongo = true
}

// EnableRedis 启用redis
func (b *Bootstrap) EnableRedis() {
	b.Redis = true
}

// Run 启动yiigo组件
func (b *Bootstrap) Run() {
	initLogger(b.Log)
	loadEnv(b.Env)
	initMySQL(b.MySQL...)

	if b.Mongo {
		initMongo()
	}

	if b.Redis {
		initRedis()
	}
}
