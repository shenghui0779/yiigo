module github.com/IIInsomnia/yiigo

require (
	github.com/go-sql-driver/mysql master
	github.com/gomodule/redigo master
	github.com/jmoiron/sqlx master
	github.com/lib/pq master
	github.com/pelletier/go-toml master
	go.uber.org/zap master
	golang.org/x/net v0.0.0
	gopkg.in/gomail.v2 master
	gopkg.in/mgo.v2 master
	gopkg.in/natefinch/lumberjack.v2 master
	vitess.io/vitess master
)

replace golang.org/x/net => github.com/golang/net master
